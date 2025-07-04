package commands

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

func (c *TelegramBotCommands) ShowCurrentListings(ctx context.Context, userID string, chatID int64) {
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

	var msgTxt string
	for idx := range allListings {
		addMsgTxt := fmt.Sprintf(fmt.Sprintf("🏠[%.0f %s %s](%s)\n%s, %s, %s\n%s\n", allListings[idx].Offers.Price, allListings[idx].Offers.PriceCurrency, escapeMarkdownV2(allListings[idx].Name), escapeMarkdownV2(allListings[idx].URL), escapeMarkdownV2(allListings[idx].Address.AddressRegion), escapeMarkdownV2(allListings[idx].Address.AddressLocality), escapeMarkdownV2(allListings[idx].Address.StreetAddress), escapeMarkdownV2(allListings[idx].CreatedAt.Format(time.RFC850))))
		if utf8.RuneCountInString(msgTxt+addMsgTxt) > messageMaxCharLen {
			c.sendMessage(chatID, userID, msgTxt, true)
			msgTxt = ""
		}
		msgTxt += addMsgTxt
	}
	if msgTxt == "" {
		msgTxt = "🤷Nothing to show, call /update_now or /run to start collecting data"
		c.sendMessage(chatID, userID, msgTxt, false)
		return
	}

	c.sendMessage(chatID, userID, msgTxt, true)
}

func escapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
