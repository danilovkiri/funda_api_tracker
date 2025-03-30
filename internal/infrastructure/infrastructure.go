package infrastructure

import (
	"context"
	"fundaNotifier/internal/infrastructure/repository/mysql"
	"sync"

	"github.com/rs/zerolog"
)

type Infrastructure struct {
	cfg       *Config
	MySqlRepo *mysql.Repository
	wg        *sync.WaitGroup
	log       *zerolog.Logger
}

func NewInfrastructure(ctx context.Context, cfg *Config, wg *sync.WaitGroup, log *zerolog.Logger) (*Infrastructure, error) {
	infrastructure := &Infrastructure{
		cfg:       cfg,
		MySqlRepo: nil,
		wg:        wg,
		log:       log,
	}
	infrastructure.mySQLInit(ctx)
	return infrastructure, nil
}

func (i *Infrastructure) mySQLInit(ctx context.Context) {
	cfg := mysql.Config{
		DNS: i.cfg.MySQLConfig.DNS,
	}
	i.MySqlRepo = mysql.NewRepository(ctx, i.wg, &cfg, i.log)
}
