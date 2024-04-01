package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type Mailer interface {
	Send(subject string, content string, to []string, cc []string, bcc []string, attachments []string) error
}

type Gmail struct {
	name             string
	fromEmailAddress string
	password         string
}

func NewGmail(name string, fromEmailAddress string, fromEmailPassword string) Mailer {
	return &Gmail{
		name:             name,
		fromEmailAddress: fromEmailAddress,
		password:         fromEmailPassword,
	}
}

func (sender *Gmail) Send(subject string, content string, to []string, cc []string, bcc []string, attachments []string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attachments {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.password, smtpAuthAddress)
	return e.Send(smtpServerAddress, smtpAuth)
}
