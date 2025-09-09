package utils

import (
	"fmt"

	"gopkg.in/mail.v2"
)

type EmailSender interface {
	SendEmail(to, subject, body string) error
}

type gomailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewGomailSender(host string, port int, username, password, from string) EmailSender {
	return &gomailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *gomailSender) SendEmail(to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(s.host, s.port, s.username, s.password)
	// d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // Use this for local testing with self-signed certs

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
