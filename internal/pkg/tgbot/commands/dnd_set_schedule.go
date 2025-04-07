package commands

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (c *TelegramBotCommands) SetDNDSchedule(ctx context.Context, userID string, chatID int64, scheduleStr string) {
	scheduleStr = strings.TrimSpace(scheduleStr)
	var schedule []string
	if scheduleStr == "" {
		schedule = []string{defaultDNDStart, defaultDNDEnd}
	} else {
		schedule = strings.Split(scheduleStr, ",")
	}

	if err := isValidTimeRange(schedule[0], schedule[1]); err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to validate DND schedule")
		msgTxt := "💥Failed to validate DND schedule"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	err := c.sessionsService.SetDNDSchedule(ctx, userID, dayTimeToMinutesAfterMidnight(schedule[0]), dayTimeToMinutesAfterMidnight(schedule[1]))
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to update DND schedule")
		msgTxt := "💥Failed to update DND schedule"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	var msgTxt string
	if scheduleStr == "" {
		msgTxt = "✅DND schedule was reset"
	} else {
		msgTxt = fmt.Sprintf("✅DND schedule was set to %s UTC — %s UTC", schedule[0], schedule[1])
	}
	c.sendMessage(chatID, userID, msgTxt)
}

func isValidTimeRange(start, end string) error {
	_, err := time.Parse(DNDLayout, start)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	_, err = time.Parse(DNDLayout, end)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	return nil
}

func dayTimeToMinutesAfterMidnight(timeStr string) int {
	t, _ := time.Parse(DNDLayout, timeStr)
	return t.Hour()*60 + t.Minute()
}
