package main

import (
	"context"
	"fundaNotifier/internal/app"
	"fundaNotifier/internal/app/bot"
	"fundaNotifier/internal/pkg/config"
	"fundaNotifier/internal/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("./.env")

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	cfg := config.NewConfig()
	log := logger.NewLog(&cfg.Logger)
	coreApp := app.New(ctx, cfg, wg, log)
	botApp := bot.New(coreApp)

	// run the app (botApp.Run is blocking, run in goroutine)
	go func() {
		if err := botApp.Run(ctx); err != nil {
			log.Error().Err(err).Msg("failed to run application, attempting shutdown")
			cancel()
			wg.Wait()
			log.Fatal().Err(err).Msg("failed to run application")
		}
	}()

	// set up an exit signal catcher
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	// graceful shutdown
	go func() {
		<-signalChan
		log.Info().Msg("application shutdown attempted")
		cancel()
	}()

	// wait for goroutines completion
	wg.Wait()
	log.Info().Msg("application shutdown finished")
}
