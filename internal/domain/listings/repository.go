package listings

import (
	"context"
	"fundaNotifier/internal/domain"
)

type Repository interface {
	Begin(ctx context.Context) (domain.Tx, error)
	MDeleteListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
	MGetListingByUserID(ctx context.Context, userID string, showOnlyNew bool) (Listings, error)
	MGetListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (Listings, error)
	GetListingByUUID(ctx context.Context, UUID string) (*Listing, error)
	MInsertListingTx(ctx context.Context, tx domain.Tx, listings Listings) error
	MUpdateListingTx(ctx context.Context, tx domain.Tx, listings Listings) error
	MDeleteListingByUserIDAndURLsTx(ctx context.Context, tx domain.Tx, userID string, URLs []string) error
	InsertFavoriteListingTx(ctx context.Context, tx domain.Tx, listing *Listing) error
	UpdateFavoriteListingTx(ctx context.Context, tx domain.Tx, listing *Listing) error
	MGetFavoriteListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (Listings, error)
	MGetFavoriteListingByUserID(ctx context.Context, userID string) (Listings, error)
	MDeleteFavoriteListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
}
