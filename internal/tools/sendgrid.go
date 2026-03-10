package tools

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridEmailSender struct {
	apiKey    string
	fromEmail string
	toEmail   string
}

func NewSendGridEmailSender(apiKey string, fromEmail string, toEmail string) *SendGridEmailSender {
	return &SendGridEmailSender{apiKey: apiKey, fromEmail: fromEmail, toEmail: toEmail}
}

func (s *SendGridEmailSender) Send(subject string, htmlBody string) error {
	from := mail.NewEmail("", s.fromEmail)
	to := mail.NewEmail("", s.toEmail)
	message := mail.NewSingleEmail(from, subject, to, "", htmlBody)
	client := sendgrid.NewSendClient(s.apiKey)
	response, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: status code %d", response.StatusCode)
	}
	return nil
}
