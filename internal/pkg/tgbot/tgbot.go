package tgbot

import (
	"context"
	"fmt"
	"fundaNotifier/internal/domain/listings"
	"fundaNotifier/internal/domain/search_queries"
	"fundaNotifier/internal/domain/sessions"
	"fundaNotifier/internal/pkg/geo"
	"fundaNotifier/internal/pkg/tgbot/commands"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

func defineCommands() tgbotapi.SetMyCommandsConfig {
	registeredCommands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Initialize the bot"},
		{Command: "run", Description: "Start scheduled API polling"},
		{Command: "pause", Description: "Pause scheduled API polling"},
		{Command: "stop", Description: "Stop the bot, remove your data and everything"},
		{Command: "set_search_query", Description: "Set search query (resets the database and current settings)"},
		{Command: "set_polling_interval", Description: "Set polling interval (e.g. `1000s`, `3m` `1h`, `1.5h`, `2h30m15s`, minimal value is 900s)"},
		{Command: "set_regions", Description: "Set regions (comma-separated, case-insensitive) or reset (if invoked without message)"},
		{Command: "set_cities", Description: "Set cities (comma-separated, case-insensitive) or reset (if invoked without message)"},
		{Command: "add_region", Description: "Add one region (case-insensitive)"},
		{Command: "add_city", Description: "Add one city (case-insensitive)"},
		{Command: "show_active_filters", Description: "Show currently set regions and cities"},
		{Command: "show_polling_interval", Description: "Show currently set polling interval"},
		{Command: "update_now", Description: "Trigger manual update"},
		{Command: "show_current_listings", Description: "Show all currently stored listings"},
		{Command: "tap_current_listings", Description: "Show all currently stored listings with an option to save any of them as favorites"},
		{Command: "show_new_listings", Description: "Show all newly added listings"},
		{Command: "tap_new_listings", Description: "Show all newly added listings with an option to save any of them as favorites"},
		{Command: "show_favorites", Description: "Show all favorite listings"},
		{Command: "dnd_set_schedule", Description: "Set DND interval in UTC as start and end HH:MM (e.g. `/set_dnd_period 23:00,08:00`), DND means that API polling will be paused during the set interval if DND is turned on"},
		{Command: "dnd_show_schedule", Description: "Show DND schedule"},
		{Command: "dnd_activate", Description: "Turn on DND"},
		{Command: "dnd_deactivate", Description: "Turn off DND"},
		{Command: "show_locations", Description: "Show a prompt with a list of major cities by their corresponding region"},
		{Command: "help", Description: "Show help"},
	}
	return tgbotapi.NewSetMyCommands(registeredCommands...)
}

type TelegramBot struct {
	log                  *zerolog.Logger
	bot                  *tgbotapi.BotAPI
	cfg                  *Config
	commands             *commands.TelegramBotCommands
	listingsService      *listings.Service
	sessionsService      *sessions.Service
	searchQueriesService *search_queries.Service
	cityData             *geo.CityData
}

func NewTelegramBot(
	cfg *Config,
	log *zerolog.Logger,
	listingsService *listings.Service,
	sessionsService *sessions.Service,
	searchQueriesService *search_queries.Service,
) *TelegramBot {
	log.Info().Msg("initializing telegram bot instance")

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("telegram bot initialization failed")
	}
	bot.Debug = true

	_, err = bot.Request(defineCommands())
	if err != nil {
		log.Fatal().Err(err).Msg("telegram bot commands initialization failed")
	}

	log.Info().Str("account", bot.Self.UserName).Msg("telegram bot was authorized as")
	return &TelegramBot{
		cfg:                  cfg,
		log:                  log,
		bot:                  bot,
		commands:             commands.NewTelegramBotCommands(log, bot, listingsService, sessionsService, searchQueriesService, geo.NewCityData()),
		listingsService:      listingsService,
		sessionsService:      sessionsService,
		searchQueriesService: searchQueriesService,
	}
}

func (b *TelegramBot) sendMessage(chatID int64, userID, message string, md2 bool) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.DisableWebPagePreview = true
	if md2 {
		msg.ParseMode = "MarkdownV2"
	}
	_, err := b.bot.Send(msg)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to send message to")
	}
}

func (b *TelegramBot) Begin(ctx context.Context, wg *sync.WaitGroup) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.runSyncerByTicker(ctx)
	}()

	for {
		select {
		case update := <-updates:
			if update.CallbackQuery != nil || update.Message.IsCommand() {
				switch {
				case update.CallbackQuery != nil:
					b.updateCallbackHandler(ctx, update)
				case update.Message.Command() != "":
					b.updateCommandHandler(ctx, update)
				default:
					break
				}
			} else {
				continue
			}
		case <-ctx.Done():
			b.log.Info().Msg("attempting to gracefully shutdown telegram bot updates handling")
			return nil
		}
	}
}

func (b *TelegramBot) updateCallbackHandler(ctx context.Context, update tgbotapi.Update) {
	user := update.CallbackQuery.From
	chatID := update.CallbackQuery.Message.Chat.ID
	msgID := update.CallbackQuery.Message.MessageID
	b.log.Debug().Str("userID", user.UserName).Int64("chatID", chatID).Msg("received callback from")

	// check that user is in whitelist
	if !b.isAuthorizedUser(user.UserName, chatID) {
		b.reactToCallbackError(user.UserName, chatID, update)
		return
	}

	if update.CallbackQuery.Data == disabledButtonCallbackData {
		b.reactToCallbackError(user.UserName, chatID, update)
		return
	}

	listing, err := b.listingsService.GetListingByUUID(ctx, update.CallbackQuery.Data)
	if err != nil {
		b.log.Error().Err(err).Str("userID", user.UserName).Msg("failed to get a listing inside a callback query")
		msgTxt := "💥Failed to get a listing inside a callback query"
		b.sendMessage(chatID, user.UserName, msgTxt, false)
		b.reactToCallbackError(user.UserName, chatID, update)
		return
	}

	err = b.listingsService.AddFavoriteListing(ctx, listing)
	if err != nil {
		b.log.Error().Err(err).Str("userID", user.UserName).Msg("failed to add a favorite listing")
		msgTxt := "💥Failed to add a favorite listing"
		b.sendMessage(chatID, user.UserName, msgTxt, false)
		b.reactToCallbackError(user.UserName, chatID, update)
		return
	}

	b.reactToCallback(user.UserName, chatID, msgID, update)
}

func (b *TelegramBot) reactToCallbackError(userID string, chatID int64, update tgbotapi.Update) {
	_, err := b.bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, "💥either an error or you have already added it to favorites"))
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to send callback call to")
	}
}

func (b *TelegramBot) reactToCallback(userID string, chatID int64, msgID int, update tgbotapi.Update) {
	_, err := b.bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, "✅"))
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to send callback call to")
	}

	updatedKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💚", disabledButtonCallbackData),
		),
	)

	// Edit the message to update the button
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, msgID, updatedKeyboard)
	if _, err = b.bot.Request(edit); err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to update button")
	}
}

func (b *TelegramBot) updateCommandHandler(ctx context.Context, update tgbotapi.Update) {
	user := update.Message.From
	chatID := update.Message.Chat.ID
	b.log.Debug().Str("userID", user.UserName).Int64("chatID", chatID).Msg("received command from")

	// check that user is in whitelist
	if !b.isAuthorizedUser(user.UserName, chatID) {
		return
	}

	// resolve command and call corresponding execution logic
	switch update.Message.Command() {
	case "start":
		if !b.canStart(ctx, user.UserName, chatID) {
			return
		}
		b.commands.Start(ctx, user.UserName, chatID)

	case "stop":
		if !b.canStop(ctx, user.UserName, chatID) {
			return
		}
		b.commands.Stop(ctx, user.UserName, chatID)

	default:
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		switch update.Message.Command() {
		case "run":
			b.commands.Run(ctx, user.UserName, chatID)

		case "pause":
			b.commands.Pause(ctx, user.UserName, chatID)

		case "set_polling_interval":
			b.commands.SetPollingInterval(ctx, user.UserName, chatID, update.Message.CommandArguments())

		case "show_polling_interval":
			b.commands.ShowPollingInterval(ctx, user.UserName, chatID)

		case "set_regions":
			b.commands.SetRegions(ctx, user.UserName, chatID, update.Message.CommandArguments())

		case "set_cities":
			b.commands.SetCities(ctx, user.UserName, chatID, update.Message.CommandArguments())

		case "add_region":
			b.commands.AddRegion(ctx, user.UserName, chatID, update.Message.CommandArguments())

		case "add_city":
			b.commands.AddCity(ctx, user.UserName, chatID, update.Message.CommandArguments())

		case "show_active_filters":
			b.commands.ShowActiveFilters(ctx, user.UserName, chatID)

		case "set_search_query":
			b.commands.SetSearchQuery(ctx, user.UserName, chatID, update.Message.CommandArguments())

		case "show_current_listings":
			b.commands.ShowCurrentListings(ctx, user.UserName, chatID)

		case "tap_current_listings":
			b.commands.TapCurrentListings(ctx, user.UserName, chatID)

		case "show_new_listings":
			b.commands.ShowNewListings(ctx, user.UserName, chatID)

		case "tap_new_listings":
			b.commands.TapNewListings(ctx, user.UserName, chatID)

		case "show_favorites":
			b.commands.ShowFavorites(ctx, user.UserName, chatID)

		case "update_now":
			b.commands.UpdateNow(ctx, user.UserName, chatID)

		case "dnd_set_schedule":
			b.commands.SetDNDSchedule(ctx, user.UserName, chatID, update.Message.CommandArguments())

		case "dnd_show_schedule":
			b.commands.ShowDNDSchedule(ctx, user.UserName, chatID)

		case "dnd_activate":
			b.commands.ActivateDND(ctx, user.UserName, chatID)

		case "dnd_deactivate":
			b.commands.DeactivateDND(ctx, user.UserName, chatID)

		case "show_locations":
			b.commands.ShowLocations(ctx, user.UserName, chatID)

		case "help":
			b.commands.Help(ctx, user.UserName, chatID)
		default:
			msgTxt := "⚠️Unknown command"
			b.sendMessage(chatID, user.UserName, msgTxt, false)
		}
	}
}

func (b *TelegramBot) runSyncerByTicker(ctx context.Context) {
	ticker := time.NewTicker(workerTickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.syncer(ctx)
		case <-ctx.Done():
			b.log.Info().Msg("shutting down in-bot ticker")
			return
		}
	}
}

func (b *TelegramBot) syncer(ctx context.Context) {
	activeSessions, err := b.sessionsService.MGetSession(ctx, true)
	if err != nil {
		b.log.Error().Err(err).Msg("failed to fetch sessions for sync")
		return
	}
	sessionsForSync := activeSessions.SelectForSync()
	b.log.Info().Int("number_of_sessions", len(sessionsForSync)).Msg("attempting to sync sessions")

	for idx := range sessionsForSync {
		if sessionsForSync[idx].IsWithinDND() {
			continue
		}
		var forceSendMessage bool
		if sessionsForSync[idx].SyncCountSinceLastChange <= nSessionsWithForcedMessageSending {
			forceSendMessage = true
		}
		b.syncerIteration(ctx, &sessionsForSync[idx], forceSendMessage)
	}
}

func (b *TelegramBot) syncerIteration(ctx context.Context, session *sessions.Session, forceSendMessage bool) {
	searchQuery, err := b.searchQueriesService.GetSearchQuery(ctx, session.UserID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to get search query for sync")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to get listings updates", time.Now().Format(time.RFC3339))
		b.sendMessage(session.ChatID, session.UserID, msgTxt, false)
		return
	}

	err = b.sessionsService.UpdateLastSyncedAt(ctx, session.UserID, time.Now())
	if err != nil {
		b.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to update last sync timestamp")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to update last sync timestamp", time.Now().Format(time.RFC3339))
		b.sendMessage(session.ChatID, session.UserID, msgTxt, false)
		return
	}

	addedListings, removedListings, _, err := b.listingsService.UpdateAndCompareListings(ctx, session.UserID, searchQuery)
	if err != nil {
		b.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to compare and update listings within sync iteration")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to get listings updates", time.Now().Format(time.RFC3339))
		b.sendMessage(session.ChatID, session.UserID, msgTxt, false)
		return
	}

	filteredAddedListings := addedListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	filteredRemovedListings := removedListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	if len(filteredAddedListings) != 0 || forceSendMessage {
		msgTxt := fmt.Sprintf("📅Updated at %s\n➕Added listings count: %d\n➖Removed listings count: %d", time.Now().Format(time.RFC3339), len(filteredAddedListings), len(filteredRemovedListings))
		b.sendMessage(session.ChatID, session.UserID, msgTxt, false)
	}
}

func escapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
