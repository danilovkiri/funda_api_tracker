package search_queries

import (
	"context"
	"fmt"
	"fundaNotifier/internal/domain"

	"github.com/rs/zerolog"
)

type Service struct {
	repository Repository
	log        *zerolog.Logger
}

func NewService(
	repository Repository,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repository: repository,
		log:        log,
	}
}

func (s *Service) GetSearchQuery(ctx context.Context, userID string) (URL string, err error) {
	URL, err = s.repository.GetSearchQueryByUserID(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to get search query")
		return URL, fmt.Errorf("failed to get search query: %w", err)
	}

	return URL, nil
}

func (s *Service) DeleteSearchQueryByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error {
	err := s.repository.DeleteSearchQueryByUserIDTx(ctx, tx, userID)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to delete search query")
		return fmt.Errorf("failed to delete search query: %w", err)
	}

	return nil
}

func (s *Service) UpsertSearchQueryByUserID(ctx context.Context, userID, searchQuery string) error {
	err := s.repository.UpsertSearchQueryByUserID(ctx, userID, searchQuery)
	if err != nil {
		s.log.Error().Err(err).Str("userID", userID).Msg("failed to upsert search query")
		return fmt.Errorf("failed to upsert search query: %w", err)
	}

	return nil
}
