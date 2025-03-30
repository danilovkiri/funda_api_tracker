package logger

type Config struct {
	Level int `env:"LOG_LEVEL" env-default:"1"`
}
