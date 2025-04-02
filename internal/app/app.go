package app

import (
	"context"
	"fundaNotifier/internal/domain/listings"
	"fundaNotifier/internal/domain/search_queries"
	"fundaNotifier/internal/domain/sessions"
	"fundaNotifier/internal/infrastructure"
	"fundaNotifier/internal/infrastructure/repository/mysql"
	"fundaNotifier/internal/integration"
	"fundaNotifier/internal/pkg/config"
	"sync"

	"github.com/rs/zerolog"
)

type Domain struct {
	Listings      *listings.Service
	SearchQueries *search_queries.Service
	Sessions      *sessions.Service
}
type App struct {
	Config            *config.Config
	Infra             *infrastructure.Infrastructure
	Integration       *integration.Integration
	Domain            *Domain
	Log               *zerolog.Logger
	Wg                *sync.WaitGroup
	ListingsRepo      *mysql.ListingsRepository
	SearchQueriesRepo *mysql.SearchQueriesRepository
	SessionsRepo      *mysql.SessionsRepository
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
	a.SearchQueriesRepo = mysql.NewSearchQueriesRepository(a.Infra.MySqlRepo)
	a.SessionsRepo = mysql.NewSessionsRepository(a.Infra.MySqlRepo)
	a.Domain.Listings = listings.NewService(a.ListingsRepo, a.Integration.FundaAPIClient, a.Log)
	a.Domain.SearchQueries = search_queries.NewService(a.SearchQueriesRepo, a.Log)
	a.Domain.Sessions = sessions.NewService(a.SessionsRepo, a.Domain.Listings, a.Domain.SearchQueries, a.Log)
}
