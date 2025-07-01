package commands

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (c *TelegramBotCommands) TapCurrentListings(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	allListings, err := c.listingsService.MGetListingByUserID(ctx, userID, false)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to get all listings")
		msgTxt := "💥Failed to get all listings"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}
	allListings = allListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	allListings.Sort()

	if len(allListings) == 0 {
		msgTxt := "🤷Nothing to show, list of favorites is empty"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	for idx := range allListings {
		msgTxt := fmt.Sprintf(fmt.Sprintf("🏠[%.0f %s %s](%s)\n%s, %s, %s\n%s\n", allListings[idx].Offers.Price, allListings[idx].Offers.PriceCurrency, escapeMarkdownV2(allListings[idx].Name), escapeMarkdownV2(allListings[idx].URL), escapeMarkdownV2(allListings[idx].Address.AddressRegion), escapeMarkdownV2(allListings[idx].Address.AddressLocality), escapeMarkdownV2(allListings[idx].Address.StreetAddress), escapeMarkdownV2(allListings[idx].CreatedAt.Format(time.RFC850))))
		rows := [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("️➕Save", allListings[idx].UUID))}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
		c.sendMessageWithKeyboard(chatID, userID, msgTxt, &keyboard, true)
	}
}
