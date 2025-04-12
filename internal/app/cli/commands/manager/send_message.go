package manager

import (
	"context"
	"fundaNotifier/internal/domain/sessions"
	"fundaNotifier/internal/pkg/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type SendMessageCommand struct {
	log             *zerolog.Logger
	sessionsService SessionsService
}

func NewSendMessageCommand(
	logger *zerolog.Logger,
	sessionsService SessionsService,
) *SendMessageCommand {
	return &SendMessageCommand{
		log:             logger,
		sessionsService: sessionsService,
	}
}

func (t *SendMessageCommand) Describe() *cli.Command {
	return &cli.Command{
		Category: "manager",
		Name:     "manager:sendMessage",
		Usage:    "Send a message to subscriber(s)",
		Action:   t.Execute,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "message",
				Usage:    "Message to send",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "userID",
				Usage:    "userID to send message to",
				Required: false,
			},
			&cli.Int64Flag{
				Name:     "chatID",
				Usage:    "chatID to send message to",
				Required: false,
			},
			&cli.BoolFlag{
				Name:  "sendToAll",
				Usage: "send message to all subscribers",
				Value: false,
			},
		},
	}
}

func (t *SendMessageCommand) Execute(ctx *cli.Context) error {
	t.log.Info().Msg("initializing telegram bot instance")
	localCtx, cancel := context.WithCancel(ctx.Context)
	defer cancel()

	_ = godotenv.Load("./.env", "./.env.local")
	cfg := config.NewConfig()
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBot.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("telegram bot initialization failed")
	}
	bot.Debug = true

	sendToAll := ctx.Bool("sendToAll")
	userID := ctx.String("userID")
	chatID := ctx.Int64("chatID")
	msgTxt := "📢" + ctx.String("message")

	var activeSessions sessions.Sessions
	if sendToAll {
		activeSessions, err = t.sessionsService.MGetSession(localCtx, true)
		if err != nil {
			t.log.Fatal().Err(err).Msg("failed to get active sessions")
		}
		for idx := range activeSessions {
			t.log.Info().Str("userID", userID).Int64("chatID", chatID).Msg("sending message to")
			t.sendMessage(bot, msgTxt, activeSessions[idx].UserID, activeSessions[idx].ChatID)
		}
	} else {
		if userID == "" || chatID == 0 {
			t.log.Fatal().Err(err).Msg("`userID` and `chatID` are required when `sendToAll` is not set")
		}
		t.log.Info().Str("userID", userID).Int64("chatID", chatID).Msg("sending message to")
		t.sendMessage(bot, msgTxt, userID, chatID)
	}

	return nil
}

func (t *SendMessageCommand) sendMessage(bot *tgbotapi.BotAPI, msgTxt string, userID string, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, msgTxt)
	_, err := bot.Send(msg)
	if err != nil {
		t.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to send message to")
	}
}
