package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"fundaNotifier/internal/domain"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type ListingsService interface {
	MDeleteListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
	MDeleteFavoriteListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error
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

func (s *Service) MGetSession(ctx context.Context, onlyActive bool) (Sessions, error) {
	activeSessions, err := s.repository.MGetSession(ctx, onlyActive)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get sessions")
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return activeSessions, nil
}

func (s *Service) SetDNDSchedule(ctx context.Context, userID string, DNDStart, DNDEnd int) error {
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

	session.DNDStart = DNDStart
	session.DNDEnd = DNDEnd

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

func (s *Service) ActivateDND(ctx context.Context, userID string) error {
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

	session.DNDActive = true

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

func (s *Service) DeactivateDND(ctx context.Context, userID string) error {
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

	session.DNDActive = false

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

func (s *Service) DeactivateSession(ctx context.Context, userID string) error {
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

	session.IsActive = false
	session.SyncCountSinceLastChange = 0

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
	session.SyncCountSinceLastChange = 0

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
	regions = strings.ToLower(regions)

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
	session.SyncCountSinceLastChange = 0

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
	cities = strings.ToLower(cities)

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
	session.SyncCountSinceLastChange = 0

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

func (s *Service) AddCity(ctx context.Context, userID string, city string) error {
	city = strings.TrimSpace(strings.ToLower(city))

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

	session.CitiesRaw += "," + city
	session.ParseRawRegionsAndCities()
	session.SyncCountSinceLastChange = 0

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

func (s *Service) AddRegion(ctx context.Context, userID string, region string) error {
	region = strings.TrimSpace(strings.ToLower(region))

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

	session.RegionsRaw += "," + region
	session.ParseRawRegionsAndCities()
	session.SyncCountSinceLastChange = 0

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
	session.SyncCountSinceLastChange += 1

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

	if err = s.listingsService.MDeleteListingByUserIDTx(ctx, tx, userID); err != nil {
		s.log.Error().Err(err).Msg("failed to delete listings upon deletion request")
		return fmt.Errorf("failed to delete listings upon deletion request: %w", err)
	}

	if err = s.listingsService.MDeleteFavoriteListingByUserIDTx(ctx, tx, userID); err != nil {
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
