package commands

import (
	"context"
	"fundaNotifier/internal/domain/listings"
	"fundaNotifier/internal/domain/sessions"
	"time"
)

type ListingsService interface {
	MGetListingByUserID(ctx context.Context, userID string, showOnlyNew bool) (listings.Listings, error)
	UpdateAndCompareListings(ctx context.Context, userID, searchQuery string) (addedListings, removedListings, leftoverListings listings.Listings, err error)
	MGetFavoriteListingByUserID(ctx context.Context, userID string) (listings.Listings, error)
}
type SessionsService interface {
	CreateDefaultSession(ctx context.Context, userID string, chatID int64) error
	GetSessionByUserID(ctx context.Context, userID string) (*sessions.Session, error)
	ActivateSession(ctx context.Context, userID string) error
	DeactivateSession(ctx context.Context, userID string) error
	UpdatePollingInterval(ctx context.Context, userID string, pollingIntervalSeconds int) error
	UpdateRegions(ctx context.Context, userID string, regions string) error
	AddRegion(ctx context.Context, userID string, region string) error
	UpdateCities(ctx context.Context, userID string, cities string) error
	AddCity(ctx context.Context, userID string, city string) error
	RemoveEverythingByUserID(ctx context.Context, userID string) error
	UpdateLastSyncedAt(ctx context.Context, userID string, lastSyncedAt time.Time) error
	SetDNDSchedule(ctx context.Context, userID string, DNDStart, DNDEnd int) error
	ActivateDND(ctx context.Context, userID string) error
	DeactivateDND(ctx context.Context, userID string) error
}
type SearchQueriesService interface {
	GetSearchQuery(ctx context.Context, userID string) (URL string, err error)
	UpsertSearchQueryByUserID(ctx context.Context, userID, searchQuery string) error
}
