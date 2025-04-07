package commands

import (
	"context"
	"strings"
)

func (c *TelegramBotCommands) ShowActiveFilters(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt)
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
	c.sendMessage(chatID, userID, msgTxt)
}
