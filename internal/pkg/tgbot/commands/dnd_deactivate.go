package commands

import "context"

func (c *TelegramBotCommands) DeactivateDND(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	if !session.DNDActive {
		msgTxt := "🤷You have already turned off the DND, there is no need to /dnd_deactivate again"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	if err = c.sessionsService.DeactivateDND(ctx, userID); err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to deactivate DND")
		msgTxt := "💥Failed to deactivate DND"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	msgTxt := "🌚You have deactivated the DND"
	c.sendMessage(chatID, userID, msgTxt, false)
}
