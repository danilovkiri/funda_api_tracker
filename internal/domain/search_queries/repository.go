package search_queries

import (
	"context"
	"fundaNotifier/internal/domain"
)

type Repository interface {
	Begin(ctx context.Context) (domain.Tx, error)
	UpsertSearchQueryByUserID(ctx context.Context, userID, searchQuery string) error
	DeleteSearchQueryByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
	GetSearchQueryByUserID(ctx context.Context, userID string) (string, error)
}
