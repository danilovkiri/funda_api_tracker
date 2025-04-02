package sessions

import (
	"context"
	"fundaNotifier/internal/domain"
)

type Repository interface {
	Begin(ctx context.Context) (domain.Tx, error)
	CreateDefaultSession(ctx context.Context, userID string, chatID int64) error
	SessionExistsByUserID(ctx context.Context, userID string) (bool, error)
	GetSessionByUserID(ctx context.Context, userID string) (*Session, error)
	GetSessionByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (*Session, error)
	UpdateSessionByUserIDTx(ctx context.Context, tx domain.Tx, session *Session) error
	DeleteSessionByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
	GetActiveSessions(ctx context.Context) (Sessions, error)
}
