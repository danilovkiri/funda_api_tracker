package listings

import (
	"context"
	"fundaNotifier/internal/domain"
)

type Repository interface {
	Begin(ctx context.Context) (domain.Tx, error)
	TruncateSearchQueryTable(ctx context.Context, tx domain.Tx) error
	TruncateListingsTable(ctx context.Context, tx domain.Tx) error
	CreateSearchQuery(ctx context.Context, URL string) error
	GetAllCurrentListings(ctx context.Context) (Listings, error)
	GetCurrentSearchQuery(ctx context.Context) (string, error)
	MUpdateListings(ctx context.Context, listings Listings) error
}
