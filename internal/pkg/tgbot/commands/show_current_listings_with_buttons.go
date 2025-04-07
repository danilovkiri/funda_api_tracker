package commands

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (c *TelegramBotCommands) ShowCurrentListingsWithButtons(ctx context.Context, userID string, chatID int64) {
	session, err := c.sessionsService.GetSessionByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to get session details")
		msgTxt := "💥Failed to get your session details"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}

	allListings, err := c.listingsService.MGetListingByUserID(ctx, userID, false)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to get all listings")
		msgTxt := "💥Failed to get all listings"
		c.sendMessage(chatID, userID, msgTxt)
		return
	}
	allListings = allListings.FilterByRegionsAndCities(session.Regions, session.Cities)
	allListings.SortByPriceDesc()

	if len(allListings) == 0 {
		msgTxt := "🤷Nothing to show, list of favorites is empty"
		c.sendMessage(chatID, userID, msgTxt)
	}

	for idx := range allListings {
		msgTxt := fmt.Sprintf(fmt.Sprintf("🏠[%.0f %s %s](%s)\n", allListings[idx].Offers.Price, allListings[idx].Offers.PriceCurrency, escapeMarkdownV2(allListings[idx].Name), escapeMarkdownV2(allListings[idx].URL)))
		rows := [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("️➕Save", allListings[idx].UUID))}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
		c.sendMessageWithKeyboard(chatID, userID, msgTxt, &keyboard)
	}
}
