package config

import (
	"fundaNotifier/internal/infrastructure"
	"fundaNotifier/internal/integration"
	"fundaNotifier/internal/pkg/logger"
	"fundaNotifier/internal/pkg/tgbot"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Infra       infrastructure.Config
	Integration integration.Config
	Logger      logger.Config
	TelegramBot tgbot.Config
}

func NewConfig() *Config {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}
