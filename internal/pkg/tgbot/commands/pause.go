package commands

import "context"

func (c *TelegramBotCommands) Pause(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	if !session.IsActive {
		msgTxt := "🤷You have already paused the polling, there is no need to /pause again"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	if err = c.sessionsService.DeactivateSession(ctx, userID); err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to deactivate polling")
		msgTxt := "💥Failed to deactivate polling"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	msgTxt := "⏸️You have paused the polling, from now on you will not receive notifications"
	c.sendMessage(chatID, userID, msgTxt)
}
