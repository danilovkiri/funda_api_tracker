package tgbot

type Config struct {
	Token           string   `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
	AuthorizedUsers []string `env:"TELEGRAM_USERS" env-required:"true"`
}
