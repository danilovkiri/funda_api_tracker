package commands

import (
	"context"
	"fmt"
	"time"
)

func (c *TelegramBotCommands) SetPollingInterval(ctx context.Context, userID string, chatID int64, duration string) {
	pollingInterval, err := time.ParseDuration(duration)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("invalid polling interval duration layout")
		msgTxt := "⚠️Invalid polling interval duration layout, valid layouts are `1000s`, `3m` `1h`, `1.5h`, `2h30m15s`, etc"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	if pollingInterval.Seconds() < minPollingIntervalSeconds {
		msgTxt := fmt.Sprintf("⚠️Polling interval cannot be less than %d seconds", minPollingIntervalSeconds)
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	err = c.sessionsService.UpdatePollingInterval(ctx, userID, int(pollingInterval.Seconds()))
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to update polling interval")
		msgTxt := "💥Failed to update polling interval"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	msgTxt := "✅Polling interval was set to: " + pollingInterval.String()
	c.sendMessage(chatID, userID, msgTxt)
}
