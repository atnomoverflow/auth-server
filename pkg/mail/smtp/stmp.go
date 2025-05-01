package smtp

import (
	"errors"
	"fmt"

	"github.com/go-gomail/gomail"

	"github.com/atnomoverflow/auth-server/pkg/mail"
)

type SMTPSender struct {
	from string
	pass string
	host string
	port uint16
}

func NewSMTPSender(from, pass, host string, port uint16) (*SMTPSender, error) {
	if !mail.IsEmailValid(from) {
		return nil, fmt.Errorf("invalid from email: %s", from)
	}

	return &SMTPSender{
		from: from,
		pass: pass,
		host: host,
		port: port,
	}, nil
}

func (s *SMTPSender) Send(input mail.SendEmailInput) error {

	if err := input.Validate(); err != nil {
		return err
	}

	msg := gomail.NewMessage()
	msg.SetHeader("Subject", input.Subject)
	msg.SetHeader("To", input.To)
	msg.SetHeader("From", s.from)
	msg.SetBody("txt/html", input.Body)
	dialer := gomail.NewDialer(s.host, int(s.port), s.from, s.pass)
	if err := dialer.DialAndSend(msg); err != nil {
		input.Logger.Error("failed to send email to:%s using smtp %w", input.To, err)
		return errors.New("fail to send email via smtp")
	}
	return nil
}
