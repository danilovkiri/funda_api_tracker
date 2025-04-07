package commands

import "context"

func (c *TelegramBotCommands) ActivateDND(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	if session.DNDActive {
		msgTxt := "🤷You have already turned on the DND, there is no need to /dnd_activate again"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	if err = c.sessionsService.ActivateDND(ctx, userID); err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to activate DND")
		msgTxt := "💥Failed to activate DND"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	msgTxt := "🌝️You have activated the DND, API polling will be disabled within the DND interval"
	c.sendMessage(chatID, userID, msgTxt, false)
}
