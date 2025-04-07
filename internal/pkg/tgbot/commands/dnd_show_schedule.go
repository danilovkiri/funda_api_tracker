package commands

import (
	"context"
	"fmt"
)

func (c *TelegramBotCommands) ShowDNDSchedule(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	var msgTxt string
	if session.DNDStart != 0 || session.DNDEnd != 0 {
		if session.DNDActive {
			msgTxt = fmt.Sprintf("🌝DND is ON\n⏳Schedule: %s UTC - %s UTC", minutesAfterMidnightToDayTime(session.DNDStart), minutesAfterMidnightToDayTime(session.DNDEnd))
		} else {
			msgTxt = fmt.Sprintf("🌚DND is OFF\n⏳Schedule: %s UTC - %s UTC", minutesAfterMidnightToDayTime(session.DNDStart), minutesAfterMidnightToDayTime(session.DNDEnd))
		}
	} else {
		msgTxt = "⚠️DND schedule is not set, run /dnd_set_schedule to set it"
	}

	c.sendMessage(chatID, userID, msgTxt)
}

func minutesAfterMidnightToDayTime(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}
