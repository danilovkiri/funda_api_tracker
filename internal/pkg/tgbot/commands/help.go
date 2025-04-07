package commands

import (
	"context"
)

func (c *TelegramBotCommands) Help(ctx context.Context, userID string, chatID int64) {
	commands, err := c.bot.GetMyCommands()
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to send message to")
		return
	}
	var msgTxt string
	for idx := range commands {
		msgTxt = msgTxt + "ℹ️/" + commands[idx].Command + " — " + commands[idx].Description + "\n"
	}
	c.sendMessage(chatID, userID, msgTxt, false)
}
