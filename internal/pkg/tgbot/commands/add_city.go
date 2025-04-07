package commands

import (
	"context"
	"fmt"
	"strings"
)

func (c *TelegramBotCommands) AddCity(ctx context.Context, userID string, chatID int64, city string) {
	city = strings.TrimSpace(city)
	if city == "" {
		c.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("failed to add empty city")
		msgTxt := "⚠️Cannot add empty city"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	err := c.sessionsService.AddCity(ctx, userID, city)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to add city")
		msgTxt := "💥Failed to add city"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	msgTxt := fmt.Sprintf("✅City %s was added", city)
	c.sendMessage(chatID, userID, msgTxt)
	c.ShowActiveFilters(ctx, userID, chatID)
}
