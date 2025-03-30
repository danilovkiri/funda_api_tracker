package tgbot

import "time"

type Config struct {
	Token                  string        `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
	AuthorizedUsers        []string      `env:"TELEGRAM_USERS" env-required:"true"`
	DefaultPollingInterval time.Duration `env:"TELEGRAM_POLLING_INTERVAL" env-default:"3600s"`
}
