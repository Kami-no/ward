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

	d := mail.NewDialer(cfg.SHost, cfg.SPort, cfg.SUser, cfg.SPass)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("Failed to send mail: %v", err)
	}

	return nil
}
