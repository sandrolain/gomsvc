package maillib

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type MailClient struct {
	From string
	addr string
	auth *smtp.Auth
}

func NewEmail() *email.Email {
	return email.NewEmail()
}

func (c *MailClient) Send(e *email.Email) error {
	if e.From == "" {
		e.From = c.From
	}
	return e.Send(c.addr, *c.auth)
}

func NewClient(cfg Config) *MailClient {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return &MailClient{
		From: cfg.From,
		addr: fmt.Sprintf("%v:%v", cfg.Host, cfg.Port),
		auth: &auth,
	}
}
