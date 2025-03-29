package infrastructure

import (
	"context"
	"fmt"
	"fundaNotifier/internal/infrastructure/repository/mysql"
	"sync"

	"github.com/rs/zerolog"
)

type Infrastructure struct {
	cfg       *Config
	mysqlRepo *mysql.Repository
	wg        *sync.WaitGroup
	log       *zerolog.Logger
}

func NewInfrastructure(ctx context.Context, cfg *Config, wg *sync.WaitGroup, log *zerolog.Logger) (*Infrastructure, error) {
	infrastructure := &Infrastructure{
		cfg:       cfg,
		mysqlRepo: nil,
		wg:        wg,
		log:       log,
	}

	if err := infrastructure.mySQLInit(ctx); err != nil {
		log.Error().Err(err).Msg("failed to initialize MySQL DB")
		return nil, fmt.Errorf("failed to initialize MySQL DB: %w", err)
	}
	return infrastructure, nil
}

func (i *Infrastructure) mySQLInit(ctx context.Context) error {
	cfg := mysql.Config{
		DNS: i.cfg.MySQLConfig.DNS,
	}
	i.mysqlRepo = mysql.NewRepository(ctx, i.wg, &cfg, i.log)
	return i.mysqlRepo.Migrate(ctx, mysql.MigrateUp)
}
