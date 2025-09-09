package utils

import (
	"fmt"
)

type mockMailSender struct {
	from string
}

func NewMockMailSender(from string) EmailSender {
	return &mockMailSender{
		from: from,
	}
}

func (s *mockMailSender) SendEmail(to, subject, body string) error {
	fmt.Printf("\n\n New Mail: \n\n From: %s \n To: %s \n Subject: %s \n Body: %s", s.from, to, subject, body)
	return nil
}
