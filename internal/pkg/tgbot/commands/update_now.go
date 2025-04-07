package commands

import (
	"context"
	"fmt"
	"time"
)

func (c *TelegramBotCommands) UpdateNow(ctx context.Context, userID string, chatID int64) {
	if !c.searchQueryIsSet(ctx, userID) {
		c.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("unable to /update_now with no search query")
		msgTxt := "💥Unable to /update_now, search query required (set it with /set_search_query followed by URL)"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	searchQuery, err := c.searchQueriesService.GetSearchQuery(ctx, session.UserID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to get search query for sync")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to get listings updates", time.Now().Format(time.RFC3339))
		c.sendMessage(session.ChatID, session.UserID, msgTxt, false)
		return
	}

	err = c.sessionsService.UpdateLastSyncedAt(ctx, session.UserID, time.Now())
	if err != nil {
		c.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to update last sync timestamp")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to update last sync timestamp", time.Now().Format(time.RFC3339))
		c.sendMessage(session.ChatID, session.UserID, msgTxt, false)
		return
	}

	addedListings, removedListings, _, err := c.listingsService.UpdateAndCompareListings(ctx, session.UserID, searchQuery)
	if err != nil {
		c.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to compare and update listings within sync iteration")
		msgTxt := fmt.Sprintf("📅Updated at %s\n💥failed to get listings updates", time.Now().Format(time.RFC3339))
		c.sendMessage(session.ChatID, session.UserID, msgTxt, false)
		return
	}

	filteredAddedListings := addedListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	filteredRemovedListings := removedListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	msgTxt := fmt.Sprintf("📅Updated at %s\n➕Added listings count: %d\n➖Removed listings count: %d", time.Now().Format(time.RFC3339), len(filteredAddedListings), len(filteredRemovedListings))
	c.sendMessage(session.ChatID, session.UserID, msgTxt, false)
}
