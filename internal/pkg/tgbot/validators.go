package tgbot

import "context"

func (b *TelegramBot) isAuthorizedUser(userID string, chatID int64) bool {
	if b.cfg.AuthorizedUsers == nil || len(b.cfg.AuthorizedUsers) == 0 {
		return true
	}

	for idx := range b.cfg.AuthorizedUsers {
		if b.cfg.AuthorizedUsers[idx] == userID {
			return true
		}
	}

	b.log.Warn().Str("userID", userID).Int64("chatID", chatID).Msg("unauthorized user detected")
	msgTxt := "🚫We are sorry, but you are not authorized to use this bot"
	b.sendMessage(chatID, userID, msgTxt)
	return false
}

func (b *TelegramBot) canStart(ctx context.Context, userID string, chatID int64) bool {
	sessionExists, err := b.sessionsService.SessionExistsByUserID(ctx, userID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to check whether session exists")
		msgTxt := "💥We failed to check whether your session already exists"
		b.sendMessage(chatID, userID, msgTxt)
		return false
	}

	if sessionExists {
		msgTxt := "🤷Your session already exists, there is nothing to /start"
		b.sendMessage(chatID, userID, msgTxt)
		return false
	}

	return true
}

func (b *TelegramBot) canStop(ctx context.Context, userID string, chatID int64) bool {
	sessionExists, err := b.sessionsService.SessionExistsByUserID(ctx, userID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to check whether session exists")
		msgTxt := "💥We failed to check whether your session already exists"
		b.sendMessage(chatID, userID, msgTxt)
		return false
	}

	if sessionExists {
		return true
	}

	msgTxt := "🤷Your session does not exist, there is nothing to /stop"
	b.sendMessage(chatID, userID, msgTxt)
	return false
}

func (b *TelegramBot) canDo(ctx context.Context, userID string, chatID int64) bool {
	sessionExists, err := b.sessionsService.SessionExistsByUserID(ctx, userID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID).Int64("chatID", chatID).Msg("failed to check whether session exists")
		msgTxt := "💥We failed to check whether your session already exists"
		b.sendMessage(chatID, userID, msgTxt)
		return false
	}

	if sessionExists {
		return true
	}

	msgTxt := "🤷Your session does not exist, run /start to initialize your session"
	b.sendMessage(chatID, userID, msgTxt)
	return false
}
