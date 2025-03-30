package bot

import (
	"context"
	"fmt"
	"fundaNotifier/internal/app"
	"fundaNotifier/internal/pkg/tgbot"
)

type Bot struct {
	*app.App
	bot *tgbot.TelegramBot
}

func New(app *app.App) *Bot {
	bot := tgbot.NewTelegramBot(&app.Config.TelegramBot, app.Log, app.Domain.Listings)
	botInstance := &Bot{
		App: app,
		bot: bot,
	}
	return botInstance
}

func (b *Bot) Run(ctx context.Context) error {
	if err := b.bot.Begin(ctx); err != nil {
		b.App.Log.Error().Err(err).Msg("failed to run bot application")
		return fmt.Errorf("failed to run bot application: %w", err)
	}
	return nil
}
