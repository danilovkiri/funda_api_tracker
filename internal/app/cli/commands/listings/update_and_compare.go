package listings

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

type UpdateAndCompareCommand struct {
	log     *zerolog.Logger
	service ListingsService
}

func NewUpdateAndCompareCommand(
	logger *zerolog.Logger,
	service ListingsService,
) *UpdateAndCompareCommand {
	return &UpdateAndCompareCommand{
		log:     logger,
		service: service,
	}
}

func (t *UpdateAndCompareCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "listings",
		Name:     "listings:updateAndCompare",
		Usage:    "Get all new listings from API, compare them with currently stored listings, show diff",
		Action:   t.Execute,
	}
}

func (t *UpdateAndCompareCommand) Execute(ctx *cli.Context) error {
	addedListings, removedListings, err := t.service.UpdateAndCompareListings(ctx.Context)
	if err != nil {
		t.log.Fatal().Err(err).Msg("failed to execute CLI command")
	}

	jsonData, err := json.MarshalIndent(addedListings, "", "  ")
	if err != nil {
		t.log.Fatal().Err(err).Msg("failed to marshall output of a CLI command")
	}
	fmt.Println("ADDED LISTINGS:\n" + string(jsonData))

	jsonData, err = json.MarshalIndent(removedListings, "", "  ")
	if err != nil {
		t.log.Fatal().Err(err).Msg("failed to marshall output of a CLI command")
	}
	fmt.Println("REMOVED LISTINGS:\n" + string(jsonData))

	return nil
}
