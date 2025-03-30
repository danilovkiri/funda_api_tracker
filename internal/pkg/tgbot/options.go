package tgbot

import "time"

type Options struct {
	PollingInterval time.Duration
	Regions         []string
	Cities          []string
	CurrentUserID   string
	CurrentChatID   int64
}

func (o *Options) Reset(pollingInterval time.Duration) {
	*o = Options{PollingInterval: pollingInterval}
}

func (o *Options) SetUserID(userID string, chatID int64) {
	o.CurrentUserID = userID
	o.CurrentChatID = chatID
}
