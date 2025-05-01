package token

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ITokenManager interface {
	Parse(string) (*jwt.Token, error)
	Sign(jwt.Claims) (string, error)
	BaseToken(string, string, time.Duration) (*BaseToken, error)
}

type JWTTokenManager struct {
	Issuer   string
	Audiance map[string]struct{}
	Key      []byte
}

const (
	ACCESS_TOKEN       = "access_token"
	ID_TOKEN           = "id_token"
	REFRESH_TOKEN      = "refresh_token"
	TWOFAUTH_TOKEN     = "2auth_token"
	VERIFICATION_TOKEN = "verification_token"
)

var (
	ErrorInvalidToken         = errors.New("invalid token")
	ErrorTokenSigningMethode  = errors.New("unexpected signing method")
	ErrorTokenInvalidIssuer   = errors.New("invalid issuer")
	ErrorTokenInvalidAudience = errors.New("invalid audience")
	ErrorTokenInvalidClaims   = errors.New("invalid claims")
	ErrorTokenExpired         = errors.New("token has expired")
)

func NewTokenManager(secret string, ops ...Options) (ITokenManager, error) {
	tokenManager := &JWTTokenManager{}

	for _, op := range ops {
		op(tokenManager)
	}
	secretEncoded, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, err
	}
	tokenManager.Key = secretEncoded
	return tokenManager, nil
}

func (tm *JWTTokenManager) Parse(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrorTokenSigningMethode
		}
		return tm.Key, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrorTokenExpired
		}
		return nil, ErrorInvalidToken
	}
	registeredClaims, ok := token.Claims.(jwt.RegisteredClaims)

	if !ok {
		return nil, ErrorTokenInvalidClaims
	}

	// Validate issuer
	if registeredClaims.Issuer != tm.Issuer {
		return nil, ErrorTokenInvalidIssuer
	}

	// Validate audience
	validAud := false
	for _, aud := range registeredClaims.Audience {
		if _, exists := tm.Audiance[aud]; exists {
			validAud = true
			break
		}
	}

	if !validAud {
		return nil, ErrorTokenInvalidAudience
	}

	return token, nil
}

func (tm *JWTTokenManager) Sign(payload jwt.Claims) (string, error) {

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// Sign the token
	signedToken, err := token.SignedString([]byte(tm.Key))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

type BaseToken struct {
	jwt.Claims
	Type string `json:"typ"`
}

func (tm *JWTTokenManager) BaseToken(subject string, tokenType string, ttl time.Duration) (*BaseToken, error) {
	audience := make([]string, 0, len(tm.Audiance))
	var i = 0
	for k := range tm.Audiance {
		audience[i] = k
		i++
	}
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return &BaseToken{
		Claims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    tm.Issuer,
			Subject:   subject,
			ID:        tokenID.String(),
			Audience:  audience,
		},
		Type: tokenType,
	}, nil
}

func ExtractToken(r *http.Request) (*string, error) {
	rawToken := r.Header.Get("Authorization")
	if rawToken == "" {
		return nil, fmt.Errorf("authorization header is missing")
	}
	parts := strings.Split(rawToken, " ")
	if len(parts) != 2 && parts[0] != "Bearer" {
		return nil, fmt.Errorf("authorion header must be in format Bearer {token}")
	}
	return &parts[1], nil
}
