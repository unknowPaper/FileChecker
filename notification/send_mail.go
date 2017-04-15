package notification

import "net/smtp"

type EmailSender interface {
	Send(to []string, title string, body string) error
}

var defaultMime = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

type EmailConfig struct {
	ServerHost string
	ServerPort string

	Username string
	Password string

	SenderAddr string
}

type emailSender struct {
	conf EmailConfig

	mime string

	send func(string, smtp.Auth, string, []string, []byte) error
}

func (e *emailSender) SetMime(newMime string) {
	e.mime = newMime
}

func (e *emailSender) Send(to []string, title string, body string) error {
	//addr := e.conf.ServerHost + ":" + e.conf.ServerPort
	//auth := smtp.PlainAuth("", e.conf.Username, e.conf.Password, e.conf.ServerHost)
	//return e.send(addr, auth, e.conf.SenderAddr, to, body)

	msg := "Subject: " + title + "\n" +
		e.mime + "\n<html><body>" +
		body + "</body></html>"

	return smtp.SendMail(e.conf.SenderAddr+":"+e.conf.ServerPort,
		smtp.PlainAuth(e.conf.Username, e.conf.SenderAddr, e.conf.Password, e.conf.ServerHost),
		e.conf.SenderAddr, to, []byte(msg))
}

func NewEmailSender(conf EmailConfig) EmailSender {
	return &emailSender{conf, defaultMime, smtp.SendMail}
}
