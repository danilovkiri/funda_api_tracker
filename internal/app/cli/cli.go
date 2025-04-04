package cli

import (
	"context"
	"fmt"
	"fundaNotifier/internal/app"
	"fundaNotifier/internal/app/cli/commands/manager"
	"fundaNotifier/internal/app/cli/commands/storage"
	"os"

	urfave "github.com/urfave/cli/v2"
)

type Cli struct {
	*app.App
	cliApp *urfave.App
}

func New(app *app.App) *Cli {
	commandMigrate := storage.NewMigrateCommand(app.Log, app.Infra.MySqlRepo)
	commandSendMessage := manager.NewSendMessageCommand(app.Log, app.Domain.Sessions)
	commandShowSessions := manager.NewShowSessionsCommand(app.Log, app.Domain.Sessions)
	commands := []*urfave.Command{
		commandMigrate.Describe(),
		commandSendMessage.Describe(),
		commandShowSessions.Describe(),
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
