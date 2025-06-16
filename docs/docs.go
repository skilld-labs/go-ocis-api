package docs

import (
	"main/rest"
)

type Config struct {
	Host     string
	Username string
	Password string
}

func New(cfg Config) *rest.Client {
	client := rest.NewClient("https", cfg.Host, "443", cfg.Username, cfg.Password, true)
	return client
}
