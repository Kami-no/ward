package main

import (
	"fmt"

	"gopkg.in/mail.v2"
)

func mailSend(cfg config, rcpt []string, subj string, msg string) error {
	m := mail.NewMessage()
	m.SetHeader("From", cfg.SMail)
	m.SetHeader("To", rcpt...)
	m.SetHeader("Subject", subj)
	m.SetBody("text/html", msg)

	d := mail.NewDialer(
		cfg.Endpoints.SMTP.Host,
		cfg.Endpoints.SMTP.Port,
		cfg.Credentials.User,
		cfg.Credentials.Password)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("Failed to send mail: %v", err)
	}

	return nil
}
