package listings

import (
	"context"
	"fundaNotifier/internal/domain/listings"
)

type ListingsService interface {
	Reset(ctx context.Context) error
	ResetAndUpdate(ctx context.Context, URL string) error
	UpdateAndCompareListings(ctx context.Context) (addedListings, removedListings listings.Listings, err error)
	GetNewListings(ctx context.Context) (listings.Listings, error)
	GetListing(ctx context.Context, URL string) (*listings.Listing, error)
	GetSearchQuery(ctx context.Context) (URL string, err error)
}
