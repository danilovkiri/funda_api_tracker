package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"fundaNotifier/internal/domain"
	tsdbmigrations "fundaNotifier/migrations"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

type Repository struct {
	cfg *Config
	db  *sql.DB
	log *zerolog.Logger
}

func NewRepository(
	ctx context.Context,
	wg *sync.WaitGroup,
	cfg *Config,
	logger *zerolog.Logger,
) *Repository {
	logger.Info().Msg("initializing MySQL DB instance")

	db, err := sql.Open("sqlite3", cfg.DNS)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to open a MySQL DB")
	}

	if err = db.Ping(); err != nil {
		logger.Fatal().Err(err).Msg("failed to ping a MySQL DB")
	}

	st := Repository{
		cfg: cfg,
		db:  db,
		log: logger,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err = st.db.Close()
		if err != nil {
			logger.Fatal().Err(err).Msg("could not close MtSQL DB")
		}
		st.db.Close()
		logger.Info().Msg("MySQL DB was closed")
	}()
	return &st
}

func (r *Repository) Migrate(ctx context.Context, direction string) error {
	goose.SetBaseFS(tsdbmigrations.EmbedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		r.log.Error().Err(err).Msg("failed to set dialect")
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	switch direction {
	case MigrateUp:
		if err := goose.Up(r.db, "."); err != nil {
			r.log.Error().Err(err).Msg("failed to migrate")
			return fmt.Errorf("failed to migrate: %w", err)
		}
	case MigrateDown:
		if err := goose.Down(r.db, "."); err != nil {
			r.log.Error().Err(err).Msg("failed to migrate")
			return fmt.Errorf("failed to migrate: %w", err)
		}
	default:
		r.log.Error().Msg("invalid migration direction key")
		return fmt.Errorf("invalid migration direction key: %s", direction)
	}
	return nil
}

func (r *Repository) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	if err := r.db.Ping(); err != nil {
		r.log.Warn().Err(err).Msg("failed to ping MySQL DB")
		return err
	}
	return nil
}

func (r *Repository) Begin(ctx context.Context) (domain.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.log.Error().Err(err).Msg("failed to begin transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}
