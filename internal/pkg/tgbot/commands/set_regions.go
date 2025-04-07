package commands

import (
	"context"
	"strings"
)

func (c *TelegramBotCommands) SetRegions(ctx context.Context, userID string, chatID int64, regions string) {
	err := c.sessionsService.UpdateRegions(ctx, userID, regions)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to update regions")
		msgTxt := "💥Failed to update regions"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	var msgTxt string
	if strings.TrimSpace(regions) == "" {
		msgTxt = "✅Regions were reset"
	} else {
		msgTxt = "✅Regions were set"
	}
	c.sendMessage(chatID, userID, msgTxt, false)
	c.ShowActiveFilters(ctx, userID, chatID)
}
