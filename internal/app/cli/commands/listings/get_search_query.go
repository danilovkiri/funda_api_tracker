package listings

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

type GetSearchQueryCommand struct {
	log     *zerolog.Logger
	service ListingsService
}

func NewGetSearchQueryCommand(
	logger *zerolog.Logger,
	service ListingsService,
) *GetSearchQueryCommand {
	return &GetSearchQueryCommand{
		log:     logger,
		service: service,
	}
}

func (t *GetSearchQueryCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "listings",
		Name:     "listings:getSearchQuery",
		Usage:    "Get parsed listing by its URL",
		Action:   t.Execute,
	}
}

func (t *GetSearchQueryCommand) Execute(ctx *cli.Context) error {
	URL, err := t.service.GetSearchQuery(ctx.Context)
	if err != nil {
		t.log.Fatal().Err(err).Msg("failed to execute CLI command")
	}

	fmt.Println(URL)

	return nil
}
