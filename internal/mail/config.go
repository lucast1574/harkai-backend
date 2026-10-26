package mail

import "net/mail"

type Config struct{ APIKey, From string }
type Result struct{ State string }

func (c Config) Valid() bool {
	address, err := mail.ParseAddress(c.From)
	return err == nil && address.Address != "" && c.APIKey != ""
}
