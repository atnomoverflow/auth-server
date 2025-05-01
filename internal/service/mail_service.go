package service

import (
	"github.com/atnomoverflow/auth-server/pkg/logger"
	"github.com/atnomoverflow/auth-server/pkg/mail"
)

type MaileService struct {
	logger    logger.ILogger
	appMailer mail.IMessage
}

func NewMailService(logger logger.ILogger, appMailer mail.IMessage) Email {
	return &MaileService{
		logger:    logger,
		appMailer: appMailer,
	}
}

func (s *MaileService) SendVerificationEmail(validationEmailInput VerificationEmailInput) error {
	// TODO: Missing the email Tempalte
	// TODO: Get the base url and concat with the verification token
	msg := mail.SendEmailInput{
		To:     validationEmailInput.To,
		Logger: s.logger,
	}

	err := s.appMailer.Send(msg)

	if err != nil {
		// TODO: Retry mechanisme
		// Might be a dead letter queue
		s.logger.Error("Mailer service failed due %w", err)
		return err
	}

	return nil
}
