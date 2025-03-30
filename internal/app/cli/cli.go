package cli

import (
	"context"
	"fmt"
	"fundaNotifier/internal/app"
	"fundaNotifier/internal/app/cli/commands/listings"
	"fundaNotifier/internal/app/cli/commands/storage"
	urfave "github.com/urfave/cli/v2"
	"os"
)

type Cli struct {
	*app.App
	cliApp *urfave.App
}

func New(app *app.App) *Cli {
	commandMigrate := storage.NewMigrateCommand(app.Log, app.Infra.MySqlRepo)
	commandReset := listings.NewResetCommand(app.Log, app.Domain.Listings)
	commandResetAndUpdate := listings.NewResetAndUpdateCommand(app.Log, app.Domain.Listings)
	commandUpdateAndCompare := listings.NewUpdateAndCompareCommand(app.Log, app.Domain.Listings)
	commandGetListing := listings.NewGetListingCommand(app.Log, app.Domain.Listings)
	commandGetNewListings := listings.NewGetNewListingsCommand(app.Log, app.Domain.Listings)
	commandgetSearchQuery := listings.NewGetSearchQueryCommand(app.Log, app.Domain.Listings)
	commands := []*urfave.Command{
		commandMigrate.Describe(),
		commandReset.Describe(),
		commandResetAndUpdate.Describe(),
		commandUpdateAndCompare.Describe(),
		commandGetListing.Describe(),
		commandGetNewListings.Describe(),
		commandgetSearchQuery.Describe(),
	}

	cliApp := &urfave.App{
		Commands: commands,
	}

	cliInstance := &Cli{
		App:    app,
		cliApp: cliApp,
	}
	return cliInstance
}

func (r *Cli) Run(ctx context.Context) error {
	if err := r.cliApp.Run(os.Args); err != nil {
		r.App.Log.Error().Err(err).Msg("failed to run CLI application")
		return fmt.Errorf("failed to run CLI application: %w", err)
	}
	return nil
}
