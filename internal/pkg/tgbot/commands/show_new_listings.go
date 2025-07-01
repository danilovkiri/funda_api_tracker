package commands

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"
)

func (c *TelegramBotCommands) ShowNewListings(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	newListings, err := c.listingsService.MGetListingByUserID(ctx, userID, true)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to get new listings")
		msgTxt := "💥Failed to get new listings"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}
	newListings = newListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	newListings.SortByPriceDesc()

	var msgTxt string
	for idx := range newListings {
		addMsgTxt := fmt.Sprintf(fmt.Sprintf("🏠[%.0f %s %s](%s)\n%s, %s, %s\n%s", newListings[idx].Offers.Price, newListings[idx].Offers.PriceCurrency, escapeMarkdownV2(newListings[idx].Name), escapeMarkdownV2(newListings[idx].URL), escapeMarkdownV2(newListings[idx].Address.AddressRegion), escapeMarkdownV2(newListings[idx].Address.AddressLocality), escapeMarkdownV2(newListings[idx].Address.StreetAddress), escapeMarkdownV2(newListings[idx].CreatedAt.Format(time.RFC850))))
		if utf8.RuneCountInString(msgTxt+addMsgTxt) > messageMaxCharLen {
			c.sendMessage(chatID, userID, msgTxt, true)
			msgTxt = ""
		}
		msgTxt += addMsgTxt
	}
	if msgTxt == "" {
		msgTxt = "🤷Nothing to show, call /update_now or /run to start collecting data; if you already did - this means that last sync retrieved zero new listings"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	c.sendMessage(chatID, userID, msgTxt, true)
}
