package commands

import (
	"context"
	"strings"
)

func (c *TelegramBotCommands) SetCities(ctx context.Context, userID string, chatID int64, cities string) {
	err := c.sessionsService.UpdateCities(ctx, userID, cities)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to update cities")
		msgTxt := "💥Failed to update cities"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	var msgTxt string
	if strings.TrimSpace(cities) == "" {
		msgTxt = "✅Cities were reset"
	} else {
		msgTxt = "✅Cities were set"
	}
	c.sendMessage(chatID, userID, msgTxt, false)
	c.ShowActiveFilters(ctx, userID, chatID)
}
