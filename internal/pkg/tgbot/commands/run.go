package commands

import "context"

func (c *TelegramBotCommands) Run(ctx context.Context, userID string, chatID int64) {
	if !c.searchQueryIsSet(ctx, userID) {
		c.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("unable to /run with no search query")
		msgTxt := "💥Unable to /run, search query required (set it with /set_search_query followed by URL)"
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

	if session.IsActive {
		msgTxt := "🤷You have already started the polling, there is no need to /run again"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	if err = c.sessionsService.ActivateSession(ctx, userID); err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to activate polling")
		msgTxt := "💥Failed to activate polling"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	msgTxt := "▶️You have started the polling, from now on you will receive notifications once per polling interval (if updates are found)"
	c.sendMessage(chatID, userID, msgTxt, false)
}

func (c *TelegramBotCommands) searchQueryIsSet(ctx context.Context, userID string) bool {
	_, err := c.searchQueriesService.GetSearchQuery(ctx, userID)
	if err != nil {
		return false
	}
	return true
}
