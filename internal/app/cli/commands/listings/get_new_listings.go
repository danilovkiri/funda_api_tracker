package listings

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

type GetNewListingsCommand struct {
	log     *zerolog.Logger
	service ListingsService
}

func NewGetNewListingsCommand(
	logger *zerolog.Logger,
	service ListingsService,
) *GetNewListingsCommand {
	return &GetNewListingsCommand{
		log:     logger,
		service: service,
	}
}

func (t *GetNewListingsCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "listings",
		Name:     "listings:getNewListings",
		Usage:    "Get all new listings from API",
		Action:   t.Execute,
	}
}

func (t *GetNewListingsCommand) Execute(ctx *cli.Context) error {
	listings, err := t.service.GetNewListings(ctx.Context)
	if err != nil {
		t.log.Fatal().Err(err).Msg("failed to execute CLI command")
	}

	jsonData, err := json.MarshalIndent(listings, "", "  ")
	if err != nil {
		t.log.Fatal().Err(err).Msg("failed to marshall output of a CLI command")
	}

	fmt.Println(string(jsonData))

	return nil
}
