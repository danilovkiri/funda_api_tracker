package listings

import (
	"context"
	"fundaNotifier/internal/domain"
)

type Repository interface {
	Begin(ctx context.Context) (domain.Tx, error)
	DeleteListingsByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
	GetListingsByUserID(ctx context.Context, userID string, showOnlyNew bool) (Listings, error)
	GetListingsByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (Listings, error)
	InsertListingsTx(ctx context.Context, tx domain.Tx, listings Listings) error
	UpdateListingsTx(ctx context.Context, tx domain.Tx, listings Listings) error
	DeleteListingsByUserIDAndURLsTx(ctx context.Context, tx domain.Tx, userID string, URLs []string) error
}
