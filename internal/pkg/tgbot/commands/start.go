package commands

import "context"

func (c *TelegramBotCommands) Start(ctx context.Context, userID string, chatID int64) {
	if err := c.sessionsService.CreateDefaultSession(ctx, userID, chatID); err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to create a new session")
		msgTxt := "💥failed to create a new session"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	msgTxt := "👋Hi\n✨Please run /help to see all available commands.\n❗You must define search query with /set_search_query\n❗You must define polling interval with /set_polling_interval\n❓You may optionally define active regions with /set_regions\n❓You may optionally define active cities with /set_cities"
	c.sendMessage(chatID, userID, msgTxt)
}
