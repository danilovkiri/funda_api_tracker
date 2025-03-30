package app

import (
	"context"
	"fundaNotifier/internal/domain/listings"
	"fundaNotifier/internal/infrastructure"
	"fundaNotifier/internal/infrastructure/repository/mysql"
	"fundaNotifier/internal/integration"
	"fundaNotifier/internal/pkg/config"
	"github.com/rs/zerolog"
	"sync"
)

type Domain struct {
	Listings *listings.Service
}
type App struct {
	Config       *config.Config
	Infra        *infrastructure.Infrastructure
	Integration  *integration.Integration
	Domain       *Domain
	Log          *zerolog.Logger
	Wg           *sync.WaitGroup
	ListingsRepo *mysql.ListingsRepository
}

func New(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup, log *zerolog.Logger) *App {
	infrastructureInstance, err := infrastructure.NewInfrastructure(ctx, &cfg.Infra, wg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start application")
	}
	integrationInstance := integration.NewIntegration(&cfg.Integration, log)
	app := &App{
		Config:      cfg,
		Infra:       infrastructureInstance,
		Integration: integrationInstance,
		Domain:      &Domain{},
		Log:         log,
		Wg:          wg,
	}
	app.Setup()
	return app
}

func (a *App) Setup() {
	a.ListingsRepo = mysql.NewListingsRepository(a.Infra.MySqlRepo)
	a.Domain.Listings = listings.NewService(a.ListingsRepo, a.Integration.FundaAPIClient, a.Log)
}
