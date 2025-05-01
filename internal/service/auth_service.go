package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/atnomoverflow/auth-server/internal/entity"
	"github.com/atnomoverflow/auth-server/internal/repository"
	"github.com/atnomoverflow/auth-server/pkg/cache"
	"github.com/atnomoverflow/auth-server/pkg/hash"
	"github.com/atnomoverflow/auth-server/pkg/logger"
	"github.com/atnomoverflow/auth-server/pkg/otp"
	"github.com/atnomoverflow/auth-server/pkg/token"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	passwordHasher         hash.Hasher
	userRepository         repository.Users
	totpManager            otp.ITOTPGenrator
	verificationCodeLength uint16
	jwtTokenManager        token.ITokenManager
	mailService            Email
	logger                 logger.ILogger
	Cache                  cache.Cache
}

const (
	TWOF_AUTH_TOKEN_TTL = time.Minute * 5
)

var (
	ErrorBadCridentials      = errors.New("bad cridentials")
	ErrorEmailUsed           = errors.New("email already exist")
	ErrorServerFailed        = errors.New("somthing went wrong pleas try again later")
	ErrorUserNotVerified     = errors.New("the user is not verified")
	ErrorUserAlreadyVerifyed = errors.New("user is already verifyed")
	// ErrorRequireTOTP2Fauth   = errors.New("require TOTP 2fauth")
	ErrorFailedTOTP2Fauth = errors.New("failed TOTP 2fauth")
	ErrorUnkownTOTP2Fauth = errors.New("unkown 2fauth methode")
	ErrorRequire2Fauth    = errors.New("require TOTP 2fauth")
)

func NewAuthService(
	passwordHasher hash.Hasher,
	userRepository repository.Users,
	verificationCodeLength uint16,
	totpManager otp.ITOTPGenrator,
	jwtTokenManager token.ITokenManager,
	logger logger.ILogger,
	mailService Email,
	Cache cache.Cache,
) *AuthService {

	return &AuthService{
		passwordHasher:         passwordHasher,
		userRepository:         userRepository,
		totpManager:            totpManager,
		jwtTokenManager:        jwtTokenManager,
		logger:                 logger,
		mailService:            mailService,
		verificationCodeLength: verificationCodeLength,
		Cache:                  Cache,
	}
}

type UserInfo struct {
	PreferredUsername string `json:"preferred_username"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
}
type AccessToken struct {
	token.BaseToken
	User UserInfo
}
type VerifycationToken struct {
	token.BaseToken
	Code string `json:"code"`
}

type TwoFAuthToken struct {
	token.BaseToken
}

func (s *AuthService) SignUp(ctx context.Context, newUser SignUpInput) error {
	existingUser, err := s.userRepository.FindByEmail(ctx, newUser.Email)

	if err != nil && err != repository.ErrorNotFound {
		s.logger.Error("failed to fetch user by email %s: %w", newUser.Email, err)
		return ErrorServerFailed
	}

	if existingUser != nil {
		s.logger.Debug("user with email %s already exists", newUser.Email)
		return ErrorEmailUsed
	}

	hashedPassword, err := s.passwordHasher.Hash(newUser.Password)

	if err != nil {
		s.logger.Error("failed to hash password for email %s: %w", newUser.Email, err)
		return ErrorServerFailed
	}

	user := &entity.User{
		FirstName:    newUser.FirstName,
		LastName:     newUser.LastName,
		Address:      newUser.Addres,
		PhoneNumber:  newUser.PhoneNumber,
		PhoneNumber2: newUser.PhoneNumber2,
		Email:        newUser.Email,
		Password:     &hashedPassword,
		IsVerified:   false,
	}

	err = s.userRepository.Create(ctx, user)

	if err != nil {
		s.logger.Error("failed to create user %s: %w", newUser.Email, err)
		return ErrorServerFailed
	}

	// send email as go corotine we dont need to wait for it to responde
	// because the user is in the database already
	go s.SendVerificationEmail(user)

	return nil
}
func (s *AuthService) SendVerificationEmail(user *entity.User) {
	baseClaims, err := s.jwtTokenManager.BaseToken(user.Email, token.VERIFICATION_TOKEN, 48*time.Hour)
	code := otp.GenrateOTP(s.verificationCodeLength)

	claims := VerifycationToken{
		BaseToken: *baseClaims,
		Code:      code,
	}

	if err != nil {
		s.logger.Error("failed to generate token for user %s: %w", user.Email, err)
		return
	}

	token, err := s.jwtTokenManager.Sign(claims)

	if err != nil {
		s.logger.Error("failed to sign token for user %s: %w", user.Email, err)
		return
	}

	err = s.mailService.SendVerificationEmail(VerificationEmailInput{
		To:    user.Email,
		Name:  fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Token: token,
	})

	if err != nil {
		s.logger.Error("failed to send verification email to user %s: %w", user.Email, err)
		return
	}
	s.logger.Info("send email to user %s: %w", user.Email, err)

}
func (s *AuthService) VerifyCredentials(ctx context.Context, credentials CredentialsInput) (*entity.User, error) {
	user, err := s.userRepository.FindByEmail(ctx, credentials.Email)

	if err != nil {
		if err == repository.ErrorNotFound {
			return nil, ErrorBadCridentials
		}
		s.logger.Error("Unknown error %v", err)
		return nil, ErrorServerFailed
	}

	isCorrect, err := s.passwordHasher.Compare(credentials.Password, *user.Password)

	if err != nil {
		s.logger.Debug("failed to compare password due %v", err)
		return nil, ErrorServerFailed
	}

	if !isCorrect {
		return nil, ErrorBadCridentials
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, credentials CredentialsInput) (*Token, *string, error) {
	user, err := s.VerifyCredentials(ctx, credentials)

	if err != nil {
		return nil, nil, err
	}

	if !user.IsVerified && time.Since(user.CreatedAt) > (48*time.Hour) {
		return nil, nil, ErrorUserNotVerified
	}

	twoFactAuth, err := s.userRepository.GetUser2Fauth(ctx, user.ID)

	if err != nil {
		if err == repository.ErrorNotFound {
			accessToken, idToken, refreshToken, err := s.generateTokens(user)
			if err != nil {
				s.logger.Error("failed to generate tokens for user %s: %v", user.Email, err)
				return nil, nil, ErrorServerFailed
			}
			s.logger.Warn("user %s don't have 2fauth enabled", user.Email)

			return &Token{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				IDToken:      idToken,
			}, nil, nil
		}
		return nil, nil, ErrorServerFailed
	}
	if len(twoFactAuth) == 0 {
		accessToken, idToken, refreshToken, err := s.generateTokens(user)
		if err != nil {
			s.logger.Error("failed to generate tokens for user %s: %v", user.Email, err)
			return nil, nil, ErrorServerFailed
		}
		return &Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			IDToken:      idToken,
		}, nil, nil
	}

	twoFactAuthToken, err := s.jwtTokenManager.BaseToken(user.Email, token.TWOFAUTH_TOKEN, TWOF_AUTH_TOKEN_TTL)
	s.logger.Warn("user %s require 2fauth", user.Email)

	if err != nil {
		s.logger.Error("failed to generate tokens for user %s: %v", user.Email, err)
		return nil, nil, ErrorServerFailed
	}

	token, err := s.jwtTokenManager.Sign(twoFactAuthToken)

	if err != nil {
		s.logger.Error("failed to generate tokens for user %s: %v", user.Email, err)
		return nil, nil, ErrorServerFailed
	}

	return nil, &token, ErrorRequire2Fauth
}

func (s *AuthService) generateTokens(user *entity.User) (string, string, string, error) {
	accessTokenClaims, err := s.jwtTokenManager.BaseToken(user.Email, token.ACCESS_TOKEN, 30*time.Minute)
	if err != nil {
		return "", "", "", err
	}
	accessToken := AccessToken{
		BaseToken: *accessTokenClaims,
		User: UserInfo{
			PreferredUsername: user.Email,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
		},
	}
	accessTokenString, err := s.jwtTokenManager.Sign(accessToken)
	if err != nil {
		return "", "", "", err
	}

	idTokenClaims, err := s.jwtTokenManager.BaseToken(user.Email, token.ID_TOKEN, 30*24*time.Hour)
	if err != nil {
		return "", "", "", err
	}
	idToken := AccessToken{
		BaseToken: *idTokenClaims,

		User: UserInfo{
			PreferredUsername: user.Email,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
		},
	}
	idTokenString, err := s.jwtTokenManager.Sign(idToken)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err := s.jwtTokenManager.BaseToken(user.Email, token.REFRESH_TOKEN, 30*24*time.Hour)
	if err != nil {
		return "", "", "", err
	}

	refreshTokenString, err := s.jwtTokenManager.Sign(refreshToken)
	if err != nil {
		return "", "", "", err
	}

	return accessTokenString, idTokenString, refreshTokenString, nil
}

func (s *AuthService) VerifyTOTP(ctx context.Context, totp TOTPVerifyInput) (*Token, error) {
	t, err := s.jwtTokenManager.Parse(totp.Token)
	if errors.Is(err, token.ErrorTokenExpired) {
		s.logger.Error("token expired %v", err)
		return nil, token.ErrorTokenExpired
	}
	if err != nil {
		s.logger.Error("token expired %v", err)
		return nil, token.ErrorInvalidToken
	}
	email, err := t.Claims.GetSubject()
	if err != nil {
		s.logger.Error("invalid email in token %v", err)
		return nil, token.ErrorInvalidToken
	}
	claims, ok := t.Claims.(token.BaseToken)

	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims.Type != token.TWOFAUTH_TOKEN {
		return nil, jwt.ErrTokenInvalidClaims
	}
	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if err == repository.ErrorNotFound {
			return nil, token.ErrorInvalidToken
		}
		return nil, ErrorServerFailed
	}
	twofauth, err := s.userRepository.GetUser2Fauth(ctx, user.ID)
	if err != nil {
		if err == repository.ErrorNotFound {
			return nil, token.ErrorInvalidToken
		}
		return nil, ErrorServerFailed
	}
	var totpConfig entity.TwoFAMethod
	for _, conf := range twofauth {
		if conf.Method == entity.TWOFAUTH_METHODE_TOTP {
			totpConfig = conf
		}
	}
	// we are asuming that the TOTPSecret is set if the methode is totp
	// finger cross
	ok = s.totpManager.Validate(string(totp.Code), *totpConfig.TOTPSecret)
	if !ok {
		return nil, ErrorFailedTOTP2Fauth
	}
	accessToken, idToken, refreshToken, err := s.generateTokens(user)
	if err != nil {
		s.logger.Error("failed to generate tokens for user %s: %v", user.Email, err)
		return nil, ErrorServerFailed
	}
	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refrechToken string) (*Token, error) {
	t, err := s.jwtTokenManager.Parse(refrechToken)
	if errors.Is(err, token.ErrorTokenExpired) {
		s.logger.Error("token expired %v", err)
		return nil, token.ErrorTokenExpired
	}
	if err != nil {
		s.logger.Error("token expired %v", err)
		return nil, token.ErrorInvalidToken
	}
	email, err := t.Claims.GetSubject()
	if err != nil {
		s.logger.Error("invalid email in token %v", err)
		return nil, token.ErrorInvalidToken
	}
	claims, ok := t.Claims.(token.BaseToken)

	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims.Type != token.REFRESH_TOKEN {
		return nil, jwt.ErrTokenInvalidClaims
	}

	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if err == repository.ErrorNotFound {
			return nil, token.ErrorInvalidToken
		}
		s.logger.Error("failed to get user by email (%s) error: %v", user.Email, err)
		return nil, ErrorServerFailed
	}
	_, err = s.Cache.Get(fmt.Sprintf("BT_%v", user.ID))
	if err == nil {
		s.logger.Error("token is revoked of user :%s", user.Email)
		return nil, token.ErrorInvalidToken
	}
	s.logger.Warn("cache missed!")
	err = s.userRepository.IsRefreshTokenBlackListed(ctx, refrechToken)
	if err != nil {
		s.logger.Error("token is revoked of user :%s", user.Email)
		return nil, token.ErrorInvalidToken
	}

	accessToken, idToken, refreshToken, err := s.generateTokens(user)
	if err != nil {
		s.logger.Error("failed to generate tokens for user %s: %v", user, err)
		return nil, ErrorServerFailed
	}
	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
	}, nil
}

func (s *AuthService) RevokeToken(ctx context.Context, refreshToken string) error {
	t, err := s.jwtTokenManager.Parse(refreshToken)
	if errors.Is(err, token.ErrorTokenExpired) {
		return token.ErrorTokenExpired
	}
	if err != nil {
		s.logger.Error("Failed to parse token: %v", err)
		return token.ErrorInvalidToken
	}
	email, err := t.Claims.GetSubject()
	if err != nil {
		s.logger.Error("Failed to get subject from token claims: %v", err)
		return token.ErrorInvalidToken
	}

	user, err := s.userRepository.FindByEmail(ctx, email)

	if err != nil {
		if err == repository.ErrorNotFound {
			return token.ErrorInvalidToken
		}
		s.logger.Error("Failed to find user by email: %v", err)
		return ErrorServerFailed
	}
	err = s.userRepository.RevokeToken(ctx, refreshToken)
	if err != nil {
		s.logger.Error("Failed revoke the token: %v", err)
		return ErrorServerFailed
	}
	go func() {
		exp, err := t.Claims.GetExpirationTime()
		if err == repository.ErrorNotFound {
			s.logger.Error("Failed to get expiration time from token claims: %v", err)
			return
		}
		currentTime := time.Now().Unix()
		expirationTime := exp.Time.Unix()
		ttl := expirationTime - currentTime

		if ttl <= 0 {
			s.logger.Warn("TTL calculation resulted in non-positive value: %d", ttl)
			return
		}

		cacheKey := fmt.Sprintf("BT_%v", user.ID)
		err = s.Cache.Set(cacheKey, refreshToken, ttl)
		if err != nil {
			s.logger.Error("Failed to set cache for key %v: %v", cacheKey, err)
		} else {
			s.logger.Info("Successfully set cache for key %v with TTL %d seconds", cacheKey, ttl)
		}
	}()
	return nil
}

// TODO: after midleware is set the user and check the sub in token
func (s *AuthService) Verify(ctx context.Context, tempToken string) error {
	t, err := s.jwtTokenManager.Parse(tempToken)
	if errors.Is(err, token.ErrorTokenExpired) {
		s.logger.Error("Failed to parse token: %v", err)
		return token.ErrorTokenExpired
	}
	if err != nil {
		return token.ErrorInvalidToken
	}
	email, err := t.Claims.GetSubject()
	if err != nil {
		s.logger.Error("Failed to get subject from token claims: %v", err)
		return token.ErrorInvalidToken
	}

	user, err := s.userRepository.FindByEmail(ctx, email)

	if err != nil {
		if err == repository.ErrorNotFound {
			return token.ErrorInvalidToken
		}
		s.logger.Error("Failed to find user by email: %v", err)
		return ErrorServerFailed
	}

	claims, ok := t.Claims.(VerifycationToken)

	if !ok {
		return jwt.ErrTokenInvalidClaims
	}

	if claims.Code != user.VerifyCode {
		return jwt.ErrTokenInvalidClaims
	}

	if user.IsVerified {
		return ErrorUserAlreadyVerifyed
	}

	user.IsVerified = true

	if err := s.userRepository.Update(ctx, user, []string{"is_verified"}); err != nil {
		return ErrorServerFailed
	}

	return nil
}
