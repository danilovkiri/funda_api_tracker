package tgbot

import (
	"context"
	"fmt"
	"fundaNotifier/internal/domain/listings"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type ListingsService interface {
	Reset(ctx context.Context) error
	ResetAndUpdate(ctx context.Context, URL string) error
	UpdateAndCompareListings(ctx context.Context) (addedListings, removedListings listings.Listings, err error)
	GetNewListings(ctx context.Context) (listings.Listings, error)
	GetListing(ctx context.Context, URL string) (*listings.Listing, error)
	GetSearchQuery(ctx context.Context) (URL string, err error)
}

type TelegramBot struct {
	log             *zerolog.Logger
	bot             *tgbotapi.BotAPI
	cfg             *Config
	opts            *Options
	listingsService ListingsService
	isActive        bool
	intervalChan    chan time.Duration
}

func NewTelegramBot(
	cfg *Config,
	log *zerolog.Logger,
	listingsService ListingsService,
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
		cfg: cfg,
		log: log,
		bot: bot,
		opts: &Options{
			PollingInterval: cfg.DefaultPollingInterval,
			Regions:         nil,
			Cities:          nil,
			CurrentUserID:   "",
			CurrentChatID:   0,
		},
		listingsService: listingsService,
		isActive:        false,
		intervalChan:    make(chan time.Duration),
	}
}

func (b *TelegramBot) Begin(ctx context.Context, wg *sync.WaitGroup) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(b.intervalChan)
		b.dynamicTicker(ctx, b.intervalChan)
	}()

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			b.updateHandler(ctx, update)
		case <-ctx.Done():
			b.log.Info().Msg("attempting to gracefully shutdown telegram bot updates handling")
			return nil
		}
	}
}

func (b *TelegramBot) updateHandler(ctx context.Context, update tgbotapi.Update) {
	user := update.Message.From
	chatID := update.Message.Chat.ID
	b.log.Debug().Str("userID", user.UserName).Int64("chatID", chatID).Msg("received message from")

	if !b.isAuthorizedUser(user.UserName, chatID) {
		return
	}

	switch update.Message.Command() {
	case "start":
		if !b.isVacant(user.UserName, chatID) {
			return
		}

		b.opts.SetUserID(user.UserName, chatID)

		msgTxt := "👋Hi\n✨Please run /help to see all available commands.\n❗You must define search query with /set_search_query\n❗You must define polling interval with /set_polling_interval\n❓You may define active regions with /set_regions\n❓You may define active cities with /set_cities"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "stop":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		err := b.listingsService.Reset(ctx)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to reset upon /stop command")
			msgTxt := "💥 failed to reset database"
			b.sendMessage(b.opts.CurrentChatID, b.opts.CurrentUserID, msgTxt)
			return
		}
		b.opts.Reset(b.cfg.DefaultPollingInterval)

		msgTxt := "✅You have stopped the bot, all your data and settings were removed"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "run":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		b.isActive = true

		msgTxt := "✅You have started the polling, from now on you will receive notifications once per polling interval (if updates are found)"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "pause":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		b.isActive = true

		msgTxt := "✅You have paused the polling, from now on you will not receive notifications"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "reset":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		err := b.listingsService.Reset(ctx)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to reset upon /reset command")
			msgTxt := "💥 failed to reset database"
			b.sendMessage(b.opts.CurrentChatID, b.opts.CurrentUserID, msgTxt)
			return
		}

		msgTxt := "✅All your data was removed from database, next sync will be run as scheduled"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "set_polling_interval":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		intervalStr := update.Message.CommandArguments()
		pollingIntervalSeconds, err := strconv.Atoi(intervalStr)
		if err != nil {
			msgTxt := "⚠️Invalid interval"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}
		msgTxt := b.setPollingInterval(pollingIntervalSeconds)
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "set_regions":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		regions := strings.Split(update.Message.CommandArguments(), ",")
		msgTxt := b.setRegions(regions)
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "set_cities":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		cities := strings.Split(update.Message.CommandArguments(), ",")
		msgTxt := b.setCities(cities)
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "set_search_query":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		searchQuery := update.Message.CommandArguments()
		if err := b.listingsService.ResetAndUpdate(ctx, searchQuery); err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to reset upon /reset command")
			msgTxt := "💥 failed to reset the database and set a new search query"
			b.sendMessage(b.opts.CurrentChatID, b.opts.CurrentUserID, msgTxt)
			return
		}

		msgTxt := "✅You have set the new search query and reset the database"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "show_active_filters":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		msgTxt := b.showRegionsAndCities()
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "show_search_query":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		searchQuery, err := b.listingsService.GetSearchQuery(ctx)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to get current search query")
			msgTxt := "💥 failed to get current search query, have you set it?"
			b.sendMessage(b.opts.CurrentChatID, b.opts.CurrentUserID, msgTxt)
			return
		}

		msgTxt := "🔗" + searchQuery
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "update_now":
		if !b.isCurrentUser(user.UserName, chatID) {
			return
		}

		addedListings, removedListings, err := b.listingsService.UpdateAndCompareListings(ctx)
		if err != nil {
			b.log.Error().Err(err).Msg("failed to run in-bot ticker handler")
			msgTxt := "💥 failed to get listings updates"
			b.sendMessage(b.opts.CurrentChatID, b.opts.CurrentUserID, msgTxt)
			return
		}
		msgTxt := fmt.Sprintf("📅Updated at %s\n➕Added listings count: %d\n➖Removed listings count: %d", time.Now().Format(time.RFC3339), len(addedListings), len(removedListings))
		b.sendMessage(b.opts.CurrentChatID, b.opts.CurrentUserID, msgTxt)

	case "help":
		commands, err := b.bot.GetMyCommands()
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to send message to")
			return
		}
		var msgTxt string
		for idx := range commands {
			msgTxt = msgTxt + "ℹ️/" + commands[idx].Command + " — " + commands[idx].Description + "\n"
		}
		b.sendMessage(chatID, user.UserName, msgTxt)

	default:
		msgTxt := "⚠️Unknown command"
		b.sendMessage(chatID, user.UserName, msgTxt)
	}
}

func (b *TelegramBot) sendMessage(chatID int64, userID, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := b.bot.Send(msg)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to send message to")
	}
}

func (b *TelegramBot) isCurrentUser(userID string, chatID int64) bool {
	if b.opts.CurrentUserID == userID && b.opts.CurrentChatID == chatID {
		return true
	}

	b.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("multiple session detected")
	msgTxt := "❌Hi. The bot is currently running for another user, who must /stop the bot before you can proceed."
	b.sendMessage(chatID, userID, msgTxt)
	return false
}

func (b *TelegramBot) isVacant(userID string, chatID int64) bool {
	if b.opts.CurrentUserID == "" && b.opts.CurrentChatID == 0 {
		return true
	}

	b.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("multiple session detected")
	msgTxt := "❌Hi. The bot is currently running for another user, which must /stop the bot for you to proceed"
	b.sendMessage(chatID, userID, msgTxt)
	return false
}

func (b *TelegramBot) isAuthorizedUser(userID string, chatID int64) bool {
	for idx := range b.cfg.AuthorizedUsers {
		if b.cfg.AuthorizedUsers[idx] == userID {
			return true
		}
	}

	b.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("unauthorized user detected")
	msgTxt := "🚫You are not authorized to use this bot"
	b.sendMessage(chatID, userID, msgTxt)
	return false
}

func (b *TelegramBot) setRegions(regions []string) string {
	b.opts.Regions = regions
	return "✅Regions were set to: " + strings.Join(b.opts.Regions, ", ")
}

func (b *TelegramBot) setCities(cities []string) string {
	b.opts.Cities = cities
	return "✅Cities were set to: " + strings.Join(b.opts.Cities, ", ")
}

func (b *TelegramBot) setPollingInterval(pollingIntervalSeconds int) string {
	pollingInterval := time.Duration(pollingIntervalSeconds) * time.Second
	b.opts.PollingInterval = pollingInterval
	b.intervalChan <- pollingInterval
	return "✅Polling interval was set to: " + b.opts.PollingInterval.String()
}

func (b *TelegramBot) showRegionsAndCities() string {
	return "🌍Active regions: " + strings.Join(b.opts.Regions, ", ") +
		"\n🌎Active cities: " + strings.Join(b.opts.Cities, ", ")
}

func (b *TelegramBot) showPollingInterval() string {
	return "⏳Active polling interval: " + b.opts.PollingInterval.String()
}

func (b *TelegramBot) dynamicTicker(ctx context.Context, intervalChan <-chan time.Duration) {
	ticker := time.NewTicker(b.cfg.DefaultPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !b.isActive {
				continue
			}
			addedListings, removedListings, err := b.listingsService.UpdateAndCompareListings(ctx)
			if err != nil {
				b.log.Error().Err(err).Msg("failed to run in-bot ticker handler")
				msgTxt := "💥 failed to get listings updates"
				b.sendMessage(b.opts.CurrentChatID, b.opts.CurrentUserID, msgTxt)
				continue
			}
			msgTxt := fmt.Sprintf("📅Updated at %s\n➕Added listings count: %d\n➖Removed listings count: %d", time.Now().Format(time.RFC3339), len(addedListings), len(removedListings))
			b.sendMessage(b.opts.CurrentChatID, b.opts.CurrentUserID, msgTxt)
		case newInterval := <-intervalChan:
			b.log.Info().Dur("new_interval", newInterval).Msg("changing in-bot ticker interval to ")
			ticker.Stop()
			ticker = time.NewTicker(newInterval)
		case <-ctx.Done():
			b.log.Info().Msg("shutting down in-bot ticker")
			return
		}
	}
}
