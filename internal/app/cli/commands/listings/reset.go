package listings

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

type ResetCommand struct {
	log     *zerolog.Logger
	service ListingsService
}

func NewResetCommand(
	logger *zerolog.Logger,
	service ListingsService,
) *ResetCommand {
	return &ResetCommand{
		log:     logger,
		service: service,
	}
}

func (t *ResetCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "listings",
		Name:     "listings:reset",
		Usage:    "Truncate all DB tables",
		Action:   t.Execute,
	}
}

func (t *ResetCommand) Execute(ctx *cli.Context) error {
	if err := t.service.Reset(ctx.Context); err != nil {
		t.log.Fatal().Err(err).Msg("failed to execute CLI command")
	}
	return nil
}
