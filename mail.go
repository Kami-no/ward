package main

import (
	"fmt"

	"gopkg.in/mail.v2"
)

func mailMergeman(cfg config, rcpt string, repo string, mr int) {
	msg := fmt.Sprintf("Hello,"+
		"<p>By merging <a href='%v'>Merge Request #%v</a> without 2 qualified approves"+
		" or negative review you've failed repository's Code of Conduct.</p>"+
		"<p>This incident will be reported.</p>", repo, mr)

	m := mail.NewMessage()
	m.SetHeader("From", cfg.SMail)
	m.SetHeader("To", rcpt)
	m.SetHeader("Subject", "Code of Conduct failure incident")
	m.SetBody("text/html", msg)

	d := mail.NewDialer(cfg.SHost, cfg.SPort, cfg.SUser, cfg.SPass)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func mailMaintainers(cfg config, rcpt []string, repo string, mr int) {
	subj := fmt.Sprintf("MR %v has failed requirements!", mr)
	msg := fmt.Sprintf("<p><a href='%v'>Merge Request #%v</a> does not meet requirements but it was merged!</p>", repo, mr)

	m := mail.NewMessage()
	m.SetHeader("From", cfg.SMail)
	m.SetHeader("To", rcpt...)
	m.SetHeader("Subject", subj)
	m.SetBody("text/html", msg)

	d := mail.NewDialer(cfg.SHost, cfg.SPort, cfg.SUser, cfg.SPass)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func mailBranches(cfg config, rcpt string, url string, branch string) {
	subj := fmt.Sprint("Dead branch detected!")
	msg := fmt.Sprintf(
		"<p>I see dead branches: <a href='%v'>%v</a></p>"+
		"<p>If you don't need it anymore, you should delete it.</p>", url, branch)

	m := mail.NewMessage()
	m.SetHeader("From", cfg.SMail)
	m.SetHeader("To", rcpt)
	m.SetHeader("Subject", subj)
	m.SetBody("text/html", msg)

	d := mail.NewDialer(cfg.SHost, cfg.SPort, cfg.SUser, cfg.SPass)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
