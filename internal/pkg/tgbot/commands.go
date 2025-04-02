package tgbot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func defineCommands() tgbotapi.SetMyCommandsConfig {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Initialize the bot"},
		{Command: "run", Description: "Start data polling and processing"},
		{Command: "pause", Description: "Pause data polling and processing"},
		{Command: "stop", Description: "Stop the bot, remove your data and everything"},
		{Command: "set_search_query", Description: "Set search query (resets the database and current settings)"},
		{Command: "set_polling_interval", Description: "Set polling interval (in seconds)"},
		{Command: "set_regions", Description: "Set regions (comma-separated)"},
		{Command: "set_cities", Description: "Set cities (comma-separated)"},
		{Command: "show_active_filters", Description: "Show currently active regions and cities"},
		{Command: "show_polling_interval", Description: "Show currently set polling interval (in seconds)"},
		{Command: "update_now", Description: "Trigger manual update"},
		{Command: "show_current_listings", Description: "Show all currently stored listings"},
		{Command: "show_new_listings", Description: "Show all newly added listings"},
		{Command: "help", Description: "Show help info"},
	}
	return tgbotapi.NewSetMyCommands(commands...)
}
