package maillib

import "github.com/jordan-wright/email"

var client *MailClient

func Configure(cfg Config) {
	client = NewClient(cfg)
}

func Send(e *email.Email) error {
	return client.Send(e)
}
