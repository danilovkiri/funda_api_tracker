package mysql

import (
	"context"
	"fmt"
	"fundaNotifier/internal/domain"
	"fundaNotifier/internal/domain/listings"
	"time"
)

var _ listings.Repository = (*ListingsRepository)(nil)

type ListingsRepository struct {
	*Repository
}

func NewListingsRepository(repository *Repository) *ListingsRepository {
	return &ListingsRepository{
		Repository: repository,
	}
}

func (r *ListingsRepository) TruncateSearchQueryTable(ctx context.Context, tx domain.Tx) error {
	const name = "ListingsRepository.TruncateSearchQueryTable"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := tx.ExecContext(ctx, "TRUNCATE listings.search_query;")
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *ListingsRepository) TruncateListingsTable(ctx context.Context, tx domain.Tx) error {
	const name = "ListingsRepository.TruncateListingsTable"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := tx.ExecContext(ctx, "TRUNCATE listings.listings;")
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *ListingsRepository) CreateSearchQuery(ctx context.Context, URL string) error {
	const name = "ListingsRepository.CreateSearchQuery"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "INSERT INTO listings.search_query (search_query, updated_at) VALUES (?, ?)", URL, time.Now())
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *ListingsRepository) GetAllCurrentListings(ctx context.Context) (listings.Listings, error) {
	//TODO implement me
	panic("implement me")
}

func (r *ListingsRepository) GetCurrentSearchQuery(ctx context.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (r *ListingsRepository) MUpdateListings(ctx context.Context, listings listings.Listings) error {
	//TODO implement me
	panic("implement me")
}
