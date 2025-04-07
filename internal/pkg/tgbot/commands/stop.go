package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (c *TelegramBotCommands) Stop(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	URL, err := c.searchQueriesService.GetSearchQuery(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			URL = "not set"
		} else {
			c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get search query upon /stop command")
			msgTxt := "💥Failed to get search query upon /stop command"
			c.sendMessage(chatID, userID, msgTxt)
			return
		}
	}

	if err = c.sessionsService.RemoveEverythingByUserID(ctx, userID); err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to remove everything upon /stop command")
		msgTxt := "💥Failed to remove everything upon /stop command"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	pollingInterval := time.Duration(session.UpdateIntervalSeconds) * time.Second
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

	msgTxt := fmt.Sprintf("⏳Polling interval: %s\n🌍Regions: %s\n📍Cities: %s\n🌝DND start: %s UTC\n🌚DND end: %s UTC\n Search query: %s\n", pollingInterval.String(), msgRegions, msgCities, minutesAfterMidnightToDayTime(session.DNDStart), minutesAfterMidnightToDayTime(session.DNDEnd), URL)
	msgTxt += "⏹️You have stopped the bot, all your data and settings were removed"
	c.sendMessage(chatID, userID, msgTxt)
}
