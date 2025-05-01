package entity

import (
	"time"

	"github.com/uptrace/bun"
)

// User represents a user entity.
type User struct {
	ID           uint64    `bun:",pk,autoincrement"`
	FirstName    string    `bun:"first_name,notnull"`
	LastName     string    `bun:"last_name,notnull"`
	Email        string    `bun:"email,unique,notnull"`
	Address      string    `bun:"address,notnull"`
	PhoneNumber  string    `bun:"phone_number"`
	PhoneNumber2 string    `bun:"phone_number_2,notnull"`
	IsVerified   bool      `bun:"is_verified,notnull,default:false"`
	VerifyCode   string    `bun:"verify_code,notnull,default:false"`
	Password     *string   `bun:",nullzero"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

// TwoFAMethod represents the 2FA method settings for a user.
type TwoFAMethod struct {
	bun.BaseModel     `bun:"table:two_fa_methods"`
	ID                int64     `bun:",pk,autoincrement"`
	UserID            uint64    `bun:"user_id,notnull"`
	Method            string    `bun:"method,notnull"`       // e.g., 'sms', 'totp'
	PhoneNumber       *string   `bun:"phone_number"`         // For SMS 2FA
	Code              *string   `bun:"code"`                 // For OTP code
	Order             uint16    `bun:"method_order,notnull"` // Order of methods
	TOTPSecret        *string   `bun:"totp_secret"`          // For TOTP
	TOTPRecoveryCodes *string   `bun:"totp_recovery_codes"`  // For TOTP recovery codes
	CreatedAt         time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt         time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

// PasswordResetToken represents a password reset token entity.
type PasswordResetToken struct {
	bun.BaseModel `bun:"table:password_reset_tokens"`
	ID            int64     `bun:",pk,autoincrement"`
	UserID        uint64    `bun:"user_id,notnull"`
	Token         string    `bun:"token,unique,notnull"`
	ExpiresAt     time.Time `bun:"expires_at,notnull"`
}

// BlackListToken represents a blacklisted token entity.
type BlackListToken struct {
	bun.BaseModel `bun:"table:black_list_tokens"`
	ID            int64  `bun:",pk,autoincrement"`
	Token         string `bun:"token,unique,notnull"`
}

type TwoFAMethodList []TwoFAMethod

func (a TwoFAMethodList) Len() int           { return len(a) }
func (a TwoFAMethodList) Less(i, j int) bool { return a[i].Order < a[j].Order }
func (a TwoFAMethodList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

const (
	TWOFAUTH_METHODE_TOTP="totp"
)