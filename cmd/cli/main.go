package main

import (
	"context"
	"fundaNotifier/internal/app"
	"fundaNotifier/internal/app/cli"
	"fundaNotifier/internal/pkg/config"
	"fundaNotifier/internal/pkg/logger"
	"github.com/joho/godotenv"
	"sync"
)

func main() {
	_ = godotenv.Load("./.env")

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	cfg := config.NewConfig()
	log := logger.NewLog(&cfg.Logger)
	coreApp := app.New(ctx, cfg, wg, log)
	cliApp := cli.New(coreApp)

	// run the app
	if err := cliApp.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to run application")
	}

	// wait for completion
	cancel()
	wg.Wait()
	log.Info().Msg("application shutdown finished")
}
