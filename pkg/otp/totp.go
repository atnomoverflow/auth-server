package otp

import "github.com/pquerna/otp/totp"

type UserOTP struct {
	Url           string
	Code          string
	RecoveryCodes []string
}

type ITOTPGenrator interface {
	Genrate(accountName string) (*UserOTP, error)
	Validate(code string, secret string) bool
}

type TOTPGenrator struct {
	Issuer             string
	RecoveryCodeCount  uint16
	RecoveryCodeLength uint16
}

func NewTotp(issuer string) *TOTPGenrator {
	return &TOTPGenrator{
		Issuer: issuer,
	}
}

func (otp TOTPGenrator) Genrate(accountName string) (*UserOTP, error) {
	userOTP := &UserOTP{}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      otp.Issuer,
		AccountName: accountName,
	})
	if err != nil {
		return nil, err
	}
	recoveryCodes := make([]string, 0, otp.RecoveryCodeCount)
	for i := 0; i < int(otp.RecoveryCodeCount); i++ {
		recoveryCodes[i] = GenrateOTP(otp.RecoveryCodeLength)
	}
	userOTP.Code = key.Secret()
	userOTP.Url = key.URL()
	userOTP.RecoveryCodes = recoveryCodes
	return userOTP, err
}
func (otp TOTPGenrator) Validate(code string, secret string) bool {
	return totp.Validate(code, secret)
}
