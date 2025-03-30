package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"fundaNotifier/internal/domain"
	"fundaNotifier/internal/domain/listings"
	"strings"
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

	_, err := tx.ExecContext(ctx, "DELETE FROM search_query;")
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

	_, err := tx.ExecContext(ctx, "DELETE FROM listings;")
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

	_, err := r.db.ExecContext(ctx, "INSERT INTO search_query (search_query, updated_at) VALUES (?, ?)", URL, time.Now())
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *ListingsRepository) GetAllCurrentListings(ctx context.Context) (listings.Listings, error) {
	const name = "ListingsRepository.GetAllCurrentListings"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	result := make(listings.Listings, 0, defaultCapacity)
	rows, err := r.db.QueryContext(ctx, "SELECT name, url, description, address_street, address_locality, address_region, currency, price, last_seen FROM listings;")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Warn().Err(err).Str("method", name).Msg("no data was found")
			return result, nil
		}
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return nil, fmt.Errorf("failed to execute query in %s: %w", name, err)
	}
	defer rows.Close()

	// iterate over rows
	for rows.Next() {
		var entry listings.Listing
		if err = rows.Scan(&entry.Name, &entry.URL, &entry.Description, &entry.Address.StreetAddress, &entry.Address.AddressLocality, &entry.Address.AddressRegion, &entry.Offers.PriceCurrency, &entry.Offers.Price, &entry.LastSeen); err != nil {
			r.log.Error().Err(err).Str("method", name).Msg("failed to scan a row in")
			return nil, fmt.Errorf("failed to scan a row in %s: %w", name, err)
		}
		result = append(result, entry)
	}
	if err = rows.Err(); err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to iterate over rows in")
		return nil, fmt.Errorf("failed to iterate over rows in %s: %w", name, err)
	}

	return result, nil
}

func (r *ListingsRepository) GetCurrentSearchQuery(ctx context.Context) (string, error) {
	const name = "ListingsRepository.GetCurrentSearchQuery"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	var searchQuery string
	err := r.db.QueryRowContext(ctx, "SELECT search_query FROM search_query;").Scan(&searchQuery)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to scan a row in")
		return searchQuery, fmt.Errorf("failed to scan a row in %s: %w", name, err)
	}

	return searchQuery, nil
}

func (r *ListingsRepository) MUpdateListings(ctx context.Context, listings listings.Listings) error {
	const fieldsLimit = 3640 // max is 32766 / 9
	if len(listings) <= fieldsLimit {
		return r.mUpdateListings(ctx, listings)
	}

	lbound := 0
	hbound := fieldsLimit
	for lbound < hbound {
		listingsSlice := listings[lbound:hbound]
		if err := r.mUpdateListings(ctx, listingsSlice); err != nil {
			return err
		}
		lbound = hbound
		hbound += fieldsLimit
		if hbound > len(listings) {
			hbound = len(listings)
		}
	}
	return nil
}

func (r *ListingsRepository) mUpdateListings(ctx context.Context, listings listings.Listings) error {
	const (
		name     = "ListingsRepository.mUpdateListings"
		fieldsNb = 9
	)
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	// build query
	b := strings.Builder{}
	params := make([]interface{}, 0, len(listings)*fieldsNb)
	b.WriteString("DELETE FROM LISTINGS; INSERT INTO listings (name, url, description, address_street, address_locality, address_region, currency, price, last_seen) VALUES ")
	counter := 0
	for idx := range listings {
		if counter > 0 {
			b.WriteString(", ")
		}
		b.WriteString("(" + strings.Repeat("?,", fieldsNb-1) + "?)")
		params = append(
			params,
			listings[idx].Name,
			listings[idx].URL,
			listings[idx].Description,
			listings[idx].Address.StreetAddress,
			listings[idx].Address.AddressLocality,
			listings[idx].Address.AddressRegion,
			listings[idx].Offers.PriceCurrency,
			listings[idx].Offers.Price,
			listings[idx].LastSeen,
		)
		counter++
	}

	_, err := r.db.ExecContext(ctx, b.String(), params...)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}
