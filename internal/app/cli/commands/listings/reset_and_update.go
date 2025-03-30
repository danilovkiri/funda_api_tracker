package listings

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

type ResetAndUpdateCommand struct {
	log     *zerolog.Logger
	service ListingsService
}

func NewResetAndUpdateCommand(
	logger *zerolog.Logger,
	service ListingsService,
) *ResetAndUpdateCommand {
	return &ResetAndUpdateCommand{
		log:     logger,
		service: service,
	}
}

func (t *ResetAndUpdateCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "listings",
		Name:     "listings:resetAndUpdate",
		Usage:    "Truncate all DB tables and set new search query",
		Action:   t.Execute,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Usage:    "New search query URL",
				Required: true,
			},
		},
	}
}

func (t *ResetAndUpdateCommand) Execute(ctx *cli.Context) error {
	URL := ctx.String("url")
	if err := t.service.ResetAndUpdate(ctx.Context, URL); err != nil {
		t.log.Fatal().Err(err).Msg("failed to execute CLI command")
	}
	return nil
}
