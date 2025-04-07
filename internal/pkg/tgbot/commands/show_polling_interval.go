package commands

import (
	"context"
	"time"
)

func (c *TelegramBotCommands) ShowPollingInterval(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	pollingInterval := time.Duration(session.UpdateIntervalSeconds) * time.Second
	msgTxt := "⏳Active polling interval: " + pollingInterval.String()
	c.sendMessage(chatID, userID, msgTxt, false)

}
