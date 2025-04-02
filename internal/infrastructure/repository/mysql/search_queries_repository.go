package mysql

import (
	"context"
	"fmt"
	"fundaNotifier/internal/domain"
	"fundaNotifier/internal/domain/search_queries"
	"time"
)

var _ search_queries.Repository = (*SearchQueriesRepository)(nil)

type SearchQueriesRepository struct {
	*Repository
}

func NewSearchQueriesRepository(repository *Repository) *SearchQueriesRepository {
	return &SearchQueriesRepository{
		Repository: repository,
	}
}

func (r *SearchQueriesRepository) UpsertSearchQueryByUserID(ctx context.Context, userID, searchQuery string) error {
	const name = "SearchQueriesRepository.UpsertSearchQueryByUserID"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "INSERT INTO search_queries (user_id, search_query) VALUES (?, ?) ON CONFLICT (user_id) DO UPDATE SET search_query = excluded.search_query;", userID, searchQuery)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *SearchQueriesRepository) DeleteSearchQueryByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error {
	const name = "SearchQueriesRepository.DeleteSearchQueryByUserID"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := tx.ExecContext(ctx, "DELETE FROM search_queries WHERE user_id = ?;", userID)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *SearchQueriesRepository) GetSearchQueryByUserID(ctx context.Context, userID string) (string, error) {
	const name = "SearchQueriesRepository.GetSearchQueryByUserID"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	var searchQuery string
	err := r.db.QueryRowContext(ctx, "SELECT search_queries FROM search_query WHERE user_id = ?;", userID).Scan(&searchQuery)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to scan a row in")
		return searchQuery, fmt.Errorf("failed to scan a row in %s: %w", name, err)
	}

	return searchQuery, nil
}
