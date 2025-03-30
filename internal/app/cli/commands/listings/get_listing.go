package listings

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

type GetListingCommand struct {
	log     *zerolog.Logger
	service ListingsService
}

func NewGetListingCommand(
	logger *zerolog.Logger,
	service ListingsService,
) *GetListingCommand {
	return &GetListingCommand{
		log:     logger,
		service: service,
	}
}

func (t *GetListingCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "listings",
		Name:     "listings:getListing",
		Usage:    "Get parsed listing by its URL",
		Action:   t.Execute,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Usage:    "Listing's URL",
				Required: true,
			},
		},
	}
}

func (t *GetListingCommand) Execute(ctx *cli.Context) error {
	URL := ctx.String("url")

	listing, err := t.service.GetListing(ctx.Context, URL)
	if err != nil {
		t.log.Fatal().Err(err).Msg("failed to execute CLI command")
	}

	jsonData, err := json.MarshalIndent(listing, "", "  ")
	if err != nil {
		t.log.Fatal().Err(err).Msg("failed to marshall output of a CLI command")
	}

	fmt.Println(string(jsonData))

	return nil
}
