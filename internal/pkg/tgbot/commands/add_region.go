package commands

import (
	"context"
	"fmt"
	"strings"
)

func (c *TelegramBotCommands) AddRegion(ctx context.Context, userID string, chatID int64, region string) {
	region = strings.TrimSpace(region)
	if region == "" {
		c.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("failed to add empty region")
		msgTxt := "⚠️Cannot add empty region"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	err := c.sessionsService.AddRegion(ctx, userID, region)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to add region")
		msgTxt := "💥Failed to add region"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	msgTxt := fmt.Sprintf("✅Region %s was added", region)
	c.sendMessage(chatID, userID, msgTxt)
	c.ShowActiveFilters(ctx, userID, chatID)
}
