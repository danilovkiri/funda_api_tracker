package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"fundaNotifier/internal/domain"
	"fundaNotifier/internal/domain/listings"
	"time"

	"github.com/rs/zerolog"
)

type ListingsService interface {
	DeleteListingsByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
	DeleteListingsByUserIDAndURLsTx(ctx context.Context, tx domain.Tx, userID string, URLs []string) error
	GetListingsByUserID(ctx context.Context, userID string, showOnlyNew bool) (listings.Listings, error)
	GetListingsByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (listings.Listings, error)
	UpsertListingsByUserIDTx(ctx context.Context, tx domain.Tx, listings listings.Listings) error
	GetCurrentlyListedListings(ctx context.Context, searchQuery string) (listings.Listings, error)
	GetListing(ctx context.Context, URL string) (*listings.Listing, error)
}

type SearchQueriesService interface {
	GetSearchQuery(ctx context.Context, userID string) (URL string, err error)
	DeleteSearchQueryByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
	UpsertSearchQueryByUserID(ctx context.Context, userID, searchQuery string) error
}

type Service struct {
	repository           Repository
	listingsService      ListingsService
	searchQueriesService SearchQueriesService
	log                  *zerolog.Logger
}

func NewService(
	repository Repository,
	listingsService ListingsService,
	searchQueriesService SearchQueriesService,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repository:           repository,
		listingsService:      listingsService,
		searchQueriesService: searchQueriesService,
		log:                  log,
	}
}

func (s *Service) CreateDefaultSession(ctx context.Context, userID string, chatID int64) error {
	err := s.repository.CreateDefaultSession(ctx, userID, chatID)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to create new default session")
		return fmt.Errorf("failed to create new default session: %w", err)
	}

	return nil
}

func (s *Service) SessionExistsByUserID(ctx context.Context, userID string) (bool, error) {
	exists, err := s.repository.SessionExistsByUserID(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to check session existence")
		return exists, fmt.Errorf("failed to check session existence: %w", err)
	}

	return exists, nil
}

func (s *Service) GetSessionByUserID(ctx context.Context, userID string) (*Session, error) {
	session, err := s.repository.GetSessionByUserID(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to get session")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (s *Service) GetSessionByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (*Session, error) {
	session, err := s.repository.GetSessionByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to get session")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}
func (s *Service) UpdateSessionByUserIDTx(ctx context.Context, tx domain.Tx, session *Session) error {
	err := s.repository.UpdateSessionByUserIDTx(ctx, tx, session)
	if err != nil {
		s.log.Error().Err(err).Str("userID", session.UserID).Msg("failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (s *Service) DeleteSessionByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error {
	err := s.repository.DeleteSessionByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (s *Service) SelectSessionsForSync(ctx context.Context) (Sessions, error) {
	activeSessions, err := s.repository.GetActiveSessions(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get active sessions")
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return activeSessions.SelectForSync(), nil
}

func (s *Service) ActivateSession(ctx context.Context, userID string) error {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to begin a transaction")
		return fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer func(tx domain.Tx) {
		errRb := tx.Rollback()
		if errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			s.log.Error().Err(errRb).Msg("failed to rollback a transaction")
		}
	}(tx)

	session, err := s.GetSessionByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get session for update")
		return fmt.Errorf("failed to get session for update: %w", err)
	}

	session.IsActive = true

	if err = s.UpdateSessionByUserIDTx(ctx, tx, session); err != nil {
		s.log.Error().Err(err).Msg("failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("failed to commit a transaction")
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return nil
}

func (s *Service) UpdatePollingInterval(ctx context.Context, userID string, pollingIntervalSeconds int) error {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to begin a transaction")
		return fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer func(tx domain.Tx) {
		errRb := tx.Rollback()
		if errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			s.log.Error().Err(errRb).Msg("failed to rollback a transaction")
		}
	}(tx)

	session, err := s.GetSessionByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get session for update")
		return fmt.Errorf("failed to get session for update: %w", err)
	}

	session.UpdateIntervalSeconds = pollingIntervalSeconds

	if err = s.UpdateSessionByUserIDTx(ctx, tx, session); err != nil {
		s.log.Error().Err(err).Msg("failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("failed to commit a transaction")
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return nil
}

func (s *Service) UpdateRegions(ctx context.Context, userID string, regions string) error {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to begin a transaction")
		return fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer func(tx domain.Tx) {
		errRb := tx.Rollback()
		if errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			s.log.Error().Err(errRb).Msg("failed to rollback a transaction")
		}
	}(tx)

	session, err := s.GetSessionByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get session for update")
		return fmt.Errorf("failed to get session for update: %w", err)
	}

	session.RegionsRaw = regions
	session.ParseRawRegionsAndCities()

	if err = s.UpdateSessionByUserIDTx(ctx, tx, session); err != nil {
		s.log.Error().Err(err).Msg("failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("failed to commit a transaction")
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return nil
}

func (s *Service) UpdateCities(ctx context.Context, userID string, cities string) error {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to begin a transaction")
		return fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer func(tx domain.Tx) {
		errRb := tx.Rollback()
		if errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			s.log.Error().Err(errRb).Msg("failed to rollback a transaction")
		}
	}(tx)

	session, err := s.GetSessionByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get session for update")
		return fmt.Errorf("failed to get session for update: %w", err)
	}

	session.CitiesRaw = cities
	session.ParseRawRegionsAndCities()

	if err = s.UpdateSessionByUserIDTx(ctx, tx, session); err != nil {
		s.log.Error().Err(err).Msg("failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("failed to commit a transaction")
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return nil
}

func (s *Service) UpdateLastSyncedAt(ctx context.Context, userID string, lastSyncedAt time.Time) error {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to begin a transaction")
		return fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer func(tx domain.Tx) {
		errRb := tx.Rollback()
		if errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			s.log.Error().Err(errRb).Msg("failed to rollback a transaction")
		}
	}(tx)

	session, err := s.GetSessionByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get session for update")
		return fmt.Errorf("failed to get session for update: %w", err)
	}

	session.LastSyncedAt = lastSyncedAt

	if err = s.UpdateSessionByUserIDTx(ctx, tx, session); err != nil {
		s.log.Error().Err(err).Msg("failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("failed to commit a transaction")
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return nil
}

func (s *Service) RemoveEverythingByUserID(ctx context.Context, userID string) error {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to begin a transaction")
		return fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer func(tx domain.Tx) {
		errRb := tx.Rollback()
		if errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			s.log.Error().Err(errRb).Msg("failed to rollback a transaction")
		}
	}(tx)

	if err = s.listingsService.DeleteListingsByUserIDTx(ctx, tx, userID); err != nil {
		s.log.Error().Err(err).Msg("failed to delete listings upon deletion request")
		return fmt.Errorf("failed to delete listings upon deletion request: %w", err)
	}

	if err = s.searchQueriesService.DeleteSearchQueryByUserIDTx(ctx, tx, userID); err != nil {
		s.log.Error().Err(err).Msg("failed to delete search query upon deletion request")
		return fmt.Errorf("failed to delete search query upon deletion request: %w", err)
	}

	if err = s.DeleteSessionByUserIDTx(ctx, tx, userID); err != nil {
		s.log.Error().Err(err).Msg("failed to delete session upon deletion request")
		return fmt.Errorf("failed to delete session upon deletion request: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("failed to commit a transaction")
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return nil
}
