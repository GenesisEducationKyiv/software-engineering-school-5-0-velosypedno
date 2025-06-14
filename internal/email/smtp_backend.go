package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPBackend struct {
	host      string
	port      string
	user      string
	pass      string
	emailFrom string
}

func NewSMTPBackend(host, port, user, pass, emailFrom string) *SMTPBackend {
	return &SMTPBackend{
		host:      host,
		port:      port,
		user:      user,
		pass:      pass,
		emailFrom: emailFrom,
	}
}

func (s *SMTPBackend) Send(to, subject, body string) error {
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.emailFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n")
	msg.WriteString(body)

	addr := s.host + ":" + s.port
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	err := smtp.SendMail(addr, auth, s.emailFrom, []string{to}, []byte(msg.String()))
	if err != nil {
		return err
	}
	return nil
}
