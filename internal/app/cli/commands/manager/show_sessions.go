package manager

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

type ShowSessionsCommand struct {
	log             *zerolog.Logger
	sessionsService SessionsService
}

func NewShowSessionsCommand(
	logger *zerolog.Logger,
	sessionsService SessionsService,
) *ShowSessionsCommand {
	return &ShowSessionsCommand{
		log:             logger,
		sessionsService: sessionsService,
	}
}

func (t *ShowSessionsCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "manager",
		Name:     "manager:showSessions",
		Usage:    "Show sessions",
		Action:   t.Execute,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "onlyActive",
				Usage: "Show only active sessions",
				Value: false,
			},
		},
	}
}

func (t *ShowSessionsCommand) Execute(ctx *cli.Context) error {
	localCtx, cancel := context.WithCancel(ctx.Context)
	defer cancel()

	onlyActive := ctx.Bool("onlyActive")

	selectedSessions, err := t.sessionsService.MGetSession(localCtx, onlyActive)
	if err != nil {
		t.log.Error().Err(err).Msg("failed to execute CLI command")
		return fmt.Errorf("failed to execute CLI command: %w", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"UserID", "ChatID", "Sync Interval", "Regions", "Cities", "Last Sync Ts", "IsActive"}
	table.SetHeader(header)
	for idx := range selectedSessions {
		var row []string
		pollingInterval := time.Duration(selectedSessions[idx].UpdateIntervalSeconds) * time.Second
		row = append(row, selectedSessions[idx].UserID, strconv.FormatInt(selectedSessions[idx].ChatID, 10), pollingInterval.String(), selectedSessions[idx].RegionsRaw, selectedSessions[idx].CitiesRaw, selectedSessions[idx].LastSyncedAt.Format(time.RFC3339), strconv.FormatBool(selectedSessions[idx].IsActive))
		table.Append(row)
	}
	table.SetAutoWrapText(false)
	table.Render()
	return nil
}
