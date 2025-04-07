package commands

import (
	"context"
	"net/url"
	"strings"
)

func (c *TelegramBotCommands) SetSearchQuery(ctx context.Context, userID string, chatID int64, searchQuery string) {
	if !validateURL(searchQuery) {
		c.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("failed to validate URL")
		msgTxt := "⚠️The provided URL is invalid, copy URL directly from browser, the URL cannot target any domain other than funda.nl"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	if err := c.searchQueriesService.UpsertSearchQueryByUserID(ctx, userID, searchQuery); err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to update search query")
		msgTxt := "💥Failed to update search query"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	msgTxt := "✅New search query was set"
	c.sendMessage(chatID, userID, msgTxt)
}

func validateURL(str string) bool {
	parsedURL, err := url.Parse(str)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if !strings.HasSuffix(parsedURL.Hostname(), "funda.nl") {
		return false
	}

	return true
}
