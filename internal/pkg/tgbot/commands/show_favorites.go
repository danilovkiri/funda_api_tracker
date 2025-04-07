package commands

import (
	"context"
	"fmt"
	"unicode/utf8"
)

func (c *TelegramBotCommands) ShowFavorites(ctx context.Context, userID string, chatID int64) {
	favorites, err := c.listingsService.MGetFavoriteListingByUserID(ctx, userID)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to get favorite listings")
		msgTxt := "💥Failed to get favorite listings"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}
	favorites.SortByPriceDesc()

	var msgTxt string
	for idx := range favorites {
		addMsgTxt := fmt.Sprintf(fmt.Sprintf("🏠[%.0f %s %s](%s)\n", favorites[idx].Offers.Price, favorites[idx].Offers.PriceCurrency, escapeMarkdownV2(favorites[idx].Name), escapeMarkdownV2(favorites[idx].URL)))
		if utf8.RuneCountInString(msgTxt+addMsgTxt) > messageMaxCharLen {
			c.sendMessage(chatID, userID, msgTxt, true)
			msgTxt = ""
		}
		msgTxt += addMsgTxt
	}
	if msgTxt == "" {
		msgTxt = "🤷Nothing to show, you need to add a favorite first"
		c.sendMessage(chatID, userID, msgTxt, false)
	}

	c.sendMessage(chatID, userID, msgTxt, true)
}
