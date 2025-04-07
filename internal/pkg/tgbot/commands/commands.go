package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type TelegramBotCommands struct {
	log                  *zerolog.Logger
	bot                  *tgbotapi.BotAPI
	listingsService      ListingsService
	sessionsService      SessionsService
	searchQueriesService SearchQueriesService
}

func NewTelegramBotCommands(
	log *zerolog.Logger,
	bot *tgbotapi.BotAPI,
	listingsService ListingsService,
	sessionsService SessionsService,
	searchQueriesService SearchQueriesService,
) *TelegramBotCommands {
	return &TelegramBotCommands{
		log:                  log,
		bot:                  bot,
		listingsService:      listingsService,
		sessionsService:      sessionsService,
		searchQueriesService: searchQueriesService,
	}
}

func (c *TelegramBotCommands) sendMessage(chatID int64, userID, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "MarkdownV2"
	_, err := c.bot.Send(msg)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to send message to")
	}
}

func (c *TelegramBotCommands) sendMessageWithKeyboard(chatID int64, userID, message string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ReplyMarkup = keyboard
	msg.DisableWebPagePreview = true
	msg.ParseMode = "MarkdownV2"
	_, err := c.bot.Send(msg)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to send message with keyboard to")
	}
}
