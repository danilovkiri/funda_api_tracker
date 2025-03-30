package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

const (
	debugLevel = iota
	infoLevel
	warnLevel
	errorLevel
)

func NewLog(cfg *Config) *zerolog.Logger {
	var level zerolog.Level
	switch cfg.Level {
	case debugLevel:
		level = zerolog.DebugLevel
	case infoLevel:
		level = zerolog.InfoLevel
	case warnLevel:
		level = zerolog.WarnLevel
	case errorLevel:
		level = zerolog.ErrorLevel
	default:
		level = zerolog.DebugLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	loggerInstance := zerolog.New(consoleWriter).With().Timestamp().Logger().Level(level)
	return &loggerInstance
}
