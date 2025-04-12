package commands

import (
	"context"
	"fmt"
)

func (c *TelegramBotCommands) ShowLocations(ctx context.Context, userID string, chatID int64) {
	regions := c.cityData.GetRegions()
	if len(regions) == 0 {
		msgTxt := "💥Nothing to show, something went wrong when loading location data"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	var msgTxt string
	for idx1 := range regions {
		msgTxt += fmt.Sprintf("🌍%s\n", regions[idx1])
		cities := c.cityData.GetCitiesByRegion(regions[idx1])
		top5Cities := cities.GetTop5ByPopulation()
		for idx2 := range top5Cities {
			msgTxt += fmt.Sprintf("\t📍%s\n", top5Cities[idx2].Name)
		}
	}
	c.sendMessage(chatID, userID, msgTxt, false)
}
