package service

import (
	"context"

	"github.com/atnomoverflow/auth-server/config"
	"github.com/atnomoverflow/auth-server/internal/repository"
	"github.com/atnomoverflow/auth-server/pkg/cache"
	"github.com/atnomoverflow/auth-server/pkg/hash"
	"github.com/atnomoverflow/auth-server/pkg/logger"
	"github.com/atnomoverflow/auth-server/pkg/mail/smtp"
	"github.com/atnomoverflow/auth-server/pkg/otp"
	"github.com/atnomoverflow/auth-server/pkg/token"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

//	type EmailVerificationInput struct {
//		Token string `json:"token"`
//	}
type SignUpInput struct {
	FirstName       string `json:"firstName" validate:"required"`
	LastName        string `json:"lastName" validate:"required"`
	Age             uint8  `json:"age" validate:"requiredgte=0,lte=130"`
	Email           string `json:"email" validate:"required,email"`
	PhoneNumber     string `json:"phoneNumber" validate:"required,e164"`
	PhoneNumber2    string `json:"phoneNumber2" validate:"required,e164"`
	Addres          string `json:"addres" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
}

type CredentialsInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TOTPVerifyInput struct {
	Token string `json:"2fa_token"validate:"required"`
	Code  uint16 `json:"code" validate:"required"`
}

//	type OTPInput struct {
//		otp string
//	}
type Auth interface {
	SignUp(ctx context.Context, newUser SignUpInput) error
	Login(ctx context.Context, cridentials CredentialsInput) (*Token, *string, error)
	VerifyTOTP(ctx context.Context, totp TOTPVerifyInput) (*Token, error)
	Verify(context.Context, string) error
	RefreshToken(ctx context.Context, refrechToken string) (*Token, error)
	RevokeToken(ctx context.Context, refrechToken string) error
}

type User interface {
	DeleteUser(ctx context.Context, id uint64) error
}

// ########### TODO: MOVE pkg

// TODO: add the payload for the emails later after creating the tempalate
type VerificationEmailInput struct {
	To    string
	Name  string
	Token string
}

type Email interface {
	//SendResetPassword(resetPasswordUrl string) error
	SendVerificationEmail(VerificationEmailInput) error

	/**
	* TODO: We might need marketing emails
	 */
}
type SMS interface {
	SendOTP(phoneNumber string, code string) error
}

type Services struct {
	EmailService Email
	UserService  User
	AuthService  Auth
}
type Dependencies struct {
	Repos                  *repository.Repositories
	PasswordHasher         hash.Hasher
	TotpManager            otp.ITOTPGenrator
	VerificationCodeLength uint16
	JWTTokenManger         token.ITokenManager
	SmtpConfig             config.Smtp
	Logger                 logger.ILogger
	Cache                  cache.Cache
}

func NewServices(deps Dependencies) *Services {
	appMailer, err := smtp.NewSMTPSender(deps.SmtpConfig.From, deps.SmtpConfig.Password, deps.SmtpConfig.Host, deps.SmtpConfig.Port)
	if err != nil {
		// TODO: a better error handling is needed!!
		panic("Invalid Email address for SMTP Server")
	}
	mailService := NewMailService(deps.Logger, appMailer)

	return &Services{
		//UserService: NewUserService(deps.Repos.User),
		AuthService: NewAuthService(
			deps.PasswordHasher,
			deps.Repos.User,
			deps.VerificationCodeLength,
			deps.TotpManager,
			deps.JWTTokenManger,
			deps.Logger,
			mailService,
			deps.Cache,
		),
	}
}
