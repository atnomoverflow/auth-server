package mail

import (
	"bytes"
	"errors"
	"html/template"

	"github.com/atnomoverflow/auth-server/pkg/logger"
)

type SendEmailInput struct {
	To      string
	Subject string
	Body    string
	Logger  logger.ILogger
}

type IMessage interface {
	Send(SendEmailInput) error
}

func (e *SendEmailInput) GenerateBodyFromHTML(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		e.Logger.Error("failed to parse file %s:%s", templateFileName, err.Error())

		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	e.Body = buf.String()

	return nil
}

func (e *SendEmailInput) Validate() error {
	if e.To == "" {
		return errors.New("empty to")
	}

	if e.Subject == "" || e.Body == "" {
		return errors.New("empty subject/body")
	}

	if !IsEmailValid(e.To) {
		return errors.New("invalid to email")
	}

	return nil
}
