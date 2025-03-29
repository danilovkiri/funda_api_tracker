package integration

import (
	"fundaNotifier/internal/integration/funda_api"
	"github.com/rs/zerolog"
)

// Integration определяет структуру объекта.
type Integration struct {
	FundaAPIClient *funda_api.FundaAPIClient
}

// NewIntegration создает экземпляр объекта Integration.
func NewIntegration(cfg *Config, log *zerolog.Logger) *Integration {
	return &Integration{
		FundaAPIClient: funda_api.NewFundaAPIClient(&cfg.FundaAPI, log),
	}
}
