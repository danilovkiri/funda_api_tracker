package storage

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

type MigrateCommand struct {
	log             *zerolog.Logger
	mySQLRepository MySQLRepository
}

func NewMigrateCommand(
	logger *zerolog.Logger,
	mySQLRepository MySQLRepository,
) *MigrateCommand {
	return &MigrateCommand{
		log:             logger,
		mySQLRepository: mySQLRepository,
	}
}

func (t *MigrateCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "storage",
		Name:     "storage:migrate",
		Usage:    "DB migration",
		Action:   t.Execute,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "direction",
				Usage:    "Migration direction, either `up` or `down`",
				Required: true,
			},
		},
	}
}

func (t *MigrateCommand) Execute(ctx *cli.Context) error {
	direction := ctx.String("direction")
	if err := t.mySQLRepository.Migrate(ctx.Context, direction); err != nil {
		t.log.Fatal().Err(err).Msg("failed to perform migration")
	}
	return nil
}
