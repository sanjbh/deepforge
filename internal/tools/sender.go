package tools

type EmailSender interface {
	Send(subject string, htmlBody string) error
}
