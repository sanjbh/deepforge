package tools

import (
	"fmt"
	"net/smtp"
)

type MailHogEmailSender struct {
	host string
	port int
	from string
	to   string
}

func NewMailHogEmailSender(host string, port int, from string, to string) *MailHogEmailSender {
	return &MailHogEmailSender{host: host, port: port, from: from, to: to}
}

func (e *MailHogEmailSender) Send(subject string, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		e.from, e.to, subject, htmlBody)
	return smtp.SendMail(addr, nil, e.from, []string{e.to}, []byte(msg))
}
