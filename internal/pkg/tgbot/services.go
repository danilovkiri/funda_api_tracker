package tgbot

import (
	"context"
	"fundaNotifier/internal/domain/listings"
	"fundaNotifier/internal/domain/sessions"
	"time"
)

type ListingsService interface {
	GetListingsByUserID(ctx context.Context, userID string, showOnlyNew bool) (listings.Listings, error)
	UpdateAndCompareListings(ctx context.Context, userID, searchQuery string) (addedListings, removedListings, leftoverListings listings.Listings, err error)
}
type SessionsService interface {
	CreateDefaultSession(ctx context.Context, userID string, chatID int64) error
	SessionExistsByUserID(ctx context.Context, userID string) (bool, error)
	GetSessionByUserID(ctx context.Context, userID string) (*sessions.Session, error)
	GetSessions(ctx context.Context, onlyActive bool) (sessions.Sessions, error)
	ActivateSession(ctx context.Context, userID string) error
	DeactivateSession(ctx context.Context, userID string) error
	UpdatePollingInterval(ctx context.Context, userID string, pollingIntervalSeconds int) error
	UpdateRegions(ctx context.Context, userID string, regions string) error
	UpdateCities(ctx context.Context, userID string, cities string) error
	RemoveEverythingByUserID(ctx context.Context, userID string) error
	UpdateLastSyncedAt(ctx context.Context, userID string, lastSyncedAt time.Time) error
}
type SearchQueryService interface {
	GetSearchQuery(ctx context.Context, userID string) (URL string, err error)
	UpsertSearchQueryByUserID(ctx context.Context, userID, searchQuery string) error
}
