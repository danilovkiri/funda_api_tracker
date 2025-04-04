package manager

import (
	"context"
	"fundaNotifier/internal/domain/sessions"
)

type SessionsService interface {
	GetSessions(ctx context.Context, onlyActive bool) (sessions.Sessions, error)
}
