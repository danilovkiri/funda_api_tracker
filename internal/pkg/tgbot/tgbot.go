package tgbot

import (
	"context"
	"fmt"
	"fundaNotifier/internal/domain"
	"fundaNotifier/internal/domain/listings"
	"fundaNotifier/internal/domain/sessions"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

const (
	messageMaxCharLen = 4096
)

type ListingsService interface {
	DeleteListingsByUserIDAndURLsTx(ctx context.Context, tx domain.Tx, userID string, URLs []string) error
	GetListingsByUserID(ctx context.Context, userID string, showOnlyNew bool) (listings.Listings, error)
	GetListingsByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (listings.Listings, error)
	GetCurrentlyListedListings(ctx context.Context, searchQuery string) (listings.Listings, error)
	UpdateAndCompareListings(ctx context.Context, userID, searchQuery string) (addedListings, removedListings, leftoverListings listings.Listings, err error)
}
type SessionsService interface {
	CreateDefaultSession(ctx context.Context, userID string, chatID int64) error
	SessionExistsByUserID(ctx context.Context, userID string) (bool, error)
	GetSessionByUserID(ctx context.Context, userID string) (*sessions.Session, error)
	SelectSessionsForSync(ctx context.Context) (sessions.Sessions, error)
	ActivateSession(ctx context.Context, userID string) error
	DeactivateSession(ctx context.Context, userID string) error
	UpdatePollingInterval(ctx context.Context, userID string, pollingIntervalSeconds int) error
	UpdateRegions(ctx context.Context, userID string, regions string) error
	UpdateCities(ctx context.Context, userID string, cities string) error
	RemoveEverythingByUserID(ctx context.Context, userID string) error
	UpdateLastSyncedAt(ctx context.Context, userID string, lastSyncedAt time.Time) error
}
type SearchQueryService interface {
	GetSearchQuery(ctx context.Context, userID string) (URL string, err error)
	UpsertSearchQueryByUserID(ctx context.Context, userID, searchQuery string) error
}

type TelegramBot struct {
	log                *zerolog.Logger
	bot                *tgbotapi.BotAPI
	cfg                *Config
	listingsService    ListingsService
	sessionsService    SessionsService
	searchQueryService SearchQueryService
}

func NewTelegramBot(
	cfg *Config,
	log *zerolog.Logger,
	listingsService ListingsService,
	sessionsService SessionsService,
	searchQueryService SearchQueryService,
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
		cfg:                cfg,
		log:                log,
		bot:                bot,
		listingsService:    listingsService,
		sessionsService:    sessionsService,
		searchQueryService: searchQueryService,
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
		if !b.canStart(ctx, user.UserName, chatID) {
			return
		}

		if err := b.sessionsService.CreateDefaultSession(ctx, user.UserName, chatID); err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to create a new session")
			msgTxt := "💥failed to create a new session"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		msgTxt := "👋Hi\n✨Please run /help to see all available commands.\n❗You must define search query with /set_search_query\n❗You must define polling interval with /set_polling_interval\n❓You may optionally define active regions with /set_regions\n❓You may optionally define active cities with /set_cities"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "stop":
		if !b.canStop(ctx, user.UserName, chatID) {
			return
		}

		if err := b.sessionsService.RemoveEverythingByUserID(ctx, user.UserName); err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to remove everything upon /stop command")
			msgTxt := "💥failed to remove everything upon /stop command"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		msgTxt := "⏹️You have stopped the bot, all your data and settings were removed"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "run":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		if !b.searchQueryIsSet(ctx, user.UserName) {
			b.log.Warn().Str("userID", user.UserName).Int64("chatID", chatID).Msg("unable to /run with no search query")
			msgTxt := "💥unable to /run, search query required (set it with /set_search_query followed by URL)"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		session, err := b.sessionsService.GetSessionByUserID(ctx, user.UserName)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to get session details")
			msgTxt := "💥failed to get your session details"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		if session.IsActive {
			msgTxt := "🤷You have already started the polling, there is no need to /run again"
			b.sendMessage(chatID, user.UserName, msgTxt)
		}

		if err = b.sessionsService.ActivateSession(ctx, user.UserName); err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to activate polling")
			msgTxt := "💥failed to activate polling"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		msgTxt := "▶️You have started the polling, from now on you will receive notifications once per polling interval (if updates are found)"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "pause":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		session, err := b.sessionsService.GetSessionByUserID(ctx, user.UserName)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to get session details")
			msgTxt := "💥failed to get your session details"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		if !session.IsActive {
			msgTxt := "🤷You have already paused the polling, there is no need to /pause again"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		if err = b.sessionsService.DeactivateSession(ctx, user.UserName); err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to deactivate polling")
			msgTxt := "💥failed to deactivate polling"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		msgTxt := "⏸️You have paused the polling, from now on you will not receive notifications"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "set_polling_interval":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		intervalStr := update.Message.CommandArguments()
		pollingIntervalSeconds, err := strconv.Atoi(intervalStr)
		if err != nil {
			msgTxt := "⚠️Invalid interval"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		err = b.sessionsService.UpdatePollingInterval(ctx, user.UserName, pollingIntervalSeconds)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to update polling interval")
			msgTxt := "💥failed to update polling interval"
			b.sendMessage(chatID, user.UserName, msgTxt)
		}

		pollingInterval := time.Duration(pollingIntervalSeconds) * time.Second
		msgTxt := "✅Polling interval was set to: " + pollingInterval.String()
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "show_polling_interval":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		session, err := b.sessionsService.GetSessionByUserID(ctx, user.UserName)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to get session details")
			msgTxt := "💥failed to get your session details"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		pollingInterval := time.Duration(session.UpdateIntervalSeconds) * time.Second
		msgTxt := "⏳Active polling interval: " + pollingInterval.String()
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "set_regions":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		regions := update.Message.CommandArguments()
		err := b.sessionsService.UpdateRegions(ctx, user.UserName, regions)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to update regions")
			msgTxt := "💥failed to update regions"
			b.sendMessage(chatID, user.UserName, msgTxt)
		}

		var msgTxt string
		if utf8.RuneCountInString(regions) == 0 {
			msgTxt = "✅Regions were reset"
		} else {
			msgTxt = "✅Regions were set"
		}
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "set_cities":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		cities := update.Message.CommandArguments()
		err := b.sessionsService.UpdateCities(ctx, user.UserName, cities)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to update cities")
			msgTxt := "💥failed to update cities"
			b.sendMessage(chatID, user.UserName, msgTxt)
		}

		var msgTxt string
		if utf8.RuneCountInString(cities) == 0 {
			msgTxt = "✅Cities were reset"
		} else {
			msgTxt = "✅Cities were set"
		}
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "show_active_filters":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		session, err := b.sessionsService.GetSessionByUserID(ctx, user.UserName)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to get session details")
			msgTxt := "💥failed to get your session details"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		var (
			msgRegions = "all"
			msgCities  = "all"
		)
		if len(session.Regions) > 0 {
			msgRegions = strings.Join(session.Regions, ", ")
		}
		if len(session.CitiesRaw) > 0 {
			msgCities = strings.Join(session.Cities, ", ")
		}
		msgTxt := "🌍Active regions: " + msgRegions +
			"\n📍Active cities: " + msgCities
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "set_search_query":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		searchQuery := update.Message.CommandArguments()
		if err := b.searchQueryService.UpsertSearchQueryByUserID(ctx, user.UserName, searchQuery); err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to update search query")
			msgTxt := "💥failed to update search query"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		msgTxt := "✅New search query was set"
		b.sendMessage(chatID, user.UserName, msgTxt)

	case "show_current_listings":
		//TODO: filter by regions and cities
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		allListings, err := b.listingsService.GetListingsByUserID(ctx, user.UserName, false)
		if err != nil {
			b.log.Error().Err(err).Msg("failed to get all listings")
			msgTxt := "💥failed to get all listings"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}
		allListings.SortByPriceDesc()

		var msgTxt string
		for idx := range allListings {
			addMsgTxt := fmt.Sprintf("%.0f %s %s\n", allListings[idx].Offers.Price, allListings[idx].Offers.PriceCurrency, allListings[idx].URL)
			if utf8.RuneCountInString(msgTxt+addMsgTxt) > messageMaxCharLen {
				b.sendMessage(chatID, user.UserName, msgTxt)
				msgTxt = ""
			}
			msgTxt += addMsgTxt
		}
		if msgTxt == "" {
			msgTxt = "🤷Nothing to show, call /update_now or /run to start collecting data"
		}

		b.sendMessage(chatID, user.UserName, msgTxt)

	case "show_new_listings":
		//TODO: filter by regions and cities
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		allListings, err := b.listingsService.GetListingsByUserID(ctx, user.UserName, true)
		if err != nil {
			b.log.Error().Err(err).Msg("failed to get all listings")
			msgTxt := "💥failed to get all listings"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		fmt.Println(allListings)

		allListings.SortByPriceDesc()

		var msgTxt string
		for idx := range allListings {
			addMsgTxt := fmt.Sprintf("%.0f %s %s\n", allListings[idx].Offers.Price, allListings[idx].Offers.PriceCurrency, allListings[idx].URL)
			if utf8.RuneCountInString(msgTxt+addMsgTxt) > messageMaxCharLen {
				b.sendMessage(chatID, user.UserName, msgTxt)
				msgTxt = ""
			}
			msgTxt += addMsgTxt
		}
		if msgTxt == "" {
			msgTxt = "🤷Nothing to show, call /update_now or /run to start collecting data; if you already did - this means that last sync retrieved zero new listings"
		}

		b.sendMessage(chatID, user.UserName, msgTxt)

	case "update_now":
		if !b.canDo(ctx, user.UserName, chatID) {
			return
		}

		if !b.searchQueryIsSet(ctx, user.UserName) {
			b.log.Warn().Str("userID", user.UserName).Int64("chatID", chatID).Msg("unable to /update_now with no search query")
			msgTxt := "💥unable to /update_now, search query required (set it with /set_search_query followed by URL)"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		session, err := b.sessionsService.GetSessionByUserID(ctx, user.UserName)
		if err != nil {
			b.log.Error().Err(err).Str("userID", user.UserName).Int64("chatID", chatID).Msg("failed to get session details")
			msgTxt := "💥failed to get your session details"
			b.sendMessage(chatID, user.UserName, msgTxt)
			return
		}

		b.syncerIteration(ctx, session, true)

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

func (b *TelegramBot) isAuthorizedUser(userID string, chatID int64) bool {
	for idx := range b.cfg.AuthorizedUsers {
		if b.cfg.AuthorizedUsers[idx] == userID {
			return true
		}
	}

	b.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("unauthorized user detected")
	msgTxt := "🚫We are sorry, but you are not authorized to use this bot"
	b.sendMessage(chatID, userID, msgTxt)
	return false
}

func (b *TelegramBot) canStart(ctx context.Context, userID string, chatID int64) bool {
	sessionExists, err := b.sessionsService.SessionExistsByUserID(ctx, userID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to check whether session exists")
		msgTxt := "💥We failed to check whether your session already exists"
		b.sendMessage(chatID, userID, msgTxt)
		return false
	}

	if sessionExists {
		msgTxt := "🤷Your session already exists, there is nothing to /start"
		b.sendMessage(chatID, userID, msgTxt)
		return false
	}

	return true
}

func (b *TelegramBot) canStop(ctx context.Context, userID string, chatID int64) bool {
	sessionExists, err := b.sessionsService.SessionExistsByUserID(ctx, userID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to check whether session exists")
		msgTxt := "💥We failed to check whether your session already exists"
		b.sendMessage(chatID, userID, msgTxt)
		return false
	}

	if sessionExists {
		return true
	}

	msgTxt := "🤷Your session does not exist, there is nothing to /stop"
	b.sendMessage(chatID, userID, msgTxt)
	return false
}

func (b *TelegramBot) canDo(ctx context.Context, userID string, chatID int64) bool {
	sessionExists, err := b.sessionsService.SessionExistsByUserID(ctx, userID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to check whether session exists")
		msgTxt := "💥We failed to check whether your session already exists"
		b.sendMessage(chatID, userID, msgTxt)
		return false
	}

	if sessionExists {
		return true
	}

	msgTxt := "❌Your session does not exist, run /start to initialize your session"
	b.sendMessage(chatID, userID, msgTxt)
	return false
}

func (b *TelegramBot) searchQueryIsSet(ctx context.Context, userID string) bool {
	_, err := b.searchQueryService.GetSearchQuery(ctx, userID)
	if err != nil {
		return false
	}
	return true
}

func (b *TelegramBot) runSyncerByTicker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.syncer(ctx, false)
		case <-ctx.Done():
			b.log.Info().Msg("shutting down in-bot ticker")
			return
		}
	}
}

func (b *TelegramBot) syncer(ctx context.Context, forcedSendMessage bool) {
	sessionsForSync, err := b.sessionsService.SelectSessionsForSync(ctx)
	if err != nil {
		b.log.Error().Err(err).Msg("failed to fetch sessions for sync")
		return
	}
	b.log.Info().Int("number_of_sessions", len(sessionsForSync)).Msg("attempting to sync sessions")

	for idx := range sessionsForSync {
		b.syncerIteration(ctx, &sessionsForSync[idx], forcedSendMessage)
	}
}

func (b *TelegramBot) syncerIteration(ctx context.Context, session *sessions.Session, forcedSendMessage bool) {
	searchQuery, err := b.searchQueryService.GetSearchQuery(ctx, session.UserID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to get search query for sync")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to get listings updates", time.Now().Format(time.RFC3339))
		b.sendMessage(session.ChatID, session.UserID, msgTxt)
		return
	}

	err = b.sessionsService.UpdateLastSyncedAt(ctx, session.UserID, time.Now())
	if err != nil {
		b.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to update last sync timestamp")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to update last sync timestamp", time.Now().Format(time.RFC3339))
		b.sendMessage(session.ChatID, session.UserID, msgTxt)
		return
	}

	addedListings, removedListings, _, err := b.listingsService.UpdateAndCompareListings(ctx, session.UserID, searchQuery)
	if err != nil {
		b.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to compare and update listings within sync iteration")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to get listings updates", time.Now().Format(time.RFC3339))
		b.sendMessage(session.ChatID, session.UserID, msgTxt)
		return
	}

	filteredAddedListings := addedListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	filteredRemovedListings := removedListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	if len(filteredAddedListings) != 0 || forcedSendMessage {
		msgTxt := fmt.Sprintf("📅Updated at %s\n➕Added listings count: %d\n➖Removed listings count: %d", time.Now().Format(time.RFC3339), len(filteredAddedListings), len(filteredRemovedListings))
		b.sendMessage(session.ChatID, session.UserID, msgTxt)
	}

}
