package otp

import "github.com/xlzd/gotp"

func GenrateOTP(legnth uint16) string {
	return gotp.RandomSecret(int(legnth))
}
