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

func (r *ListingsRepository) MDeleteListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error {
	const name = "ListingsRepository.MDeleteListingByUserIDTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := tx.ExecContext(ctx, "DELETE FROM listings WHERE user_id = ?;", userID)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *ListingsRepository) MDeleteListingByUserIDAndURLsTx(ctx context.Context, tx domain.Tx, userID string, URLs []string) error {
	const name = "ListingsRepository.MDeleteListingByUserIDAndURLsTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	stmt, err := tx.PrepareContext(ctx, "DELETE FROM listings WHERE user_id = ? AND url = ?;")
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to prepare statement in")
		return fmt.Errorf("failed to prepare statement in %s: %w", name, err)
	}
	defer stmt.Close()

	for idx := range URLs {
		_, err = stmt.ExecContext(ctx, userID, URLs[idx])
		if err != nil {
			r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
			return fmt.Errorf("failed to execute query in %s: %w", name, err)
		}
	}

	return nil
}

func (r *ListingsRepository) GetListingByUUID(ctx context.Context, UUID string) (*listings.Listing, error) {
	const name = "ListingsRepository.GetListingByUUID"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	var entry listings.Listing
	err := r.db.QueryRowContext(ctx, "SELECT user_id, name, url, description, address_street, address_locality, address_region, currency, price, is_new, uuid FROM listings WHERE uuid = ?;", UUID).Scan(&entry.UserID, &entry.Name, &entry.URL, &entry.Description, &entry.Address.StreetAddress, &entry.Address.AddressLocality, &entry.Address.AddressRegion, &entry.Offers.PriceCurrency, &entry.Offers.Price, &entry.IsNew, &entry.UUID)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return nil, fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return &entry, nil
}

func (r *ListingsRepository) MGetListingByUserID(ctx context.Context, userID string, showOnlyNew bool) (listings.Listings, error) {
	const name = "ListingsRepository.MGetListingByUserID"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	var query string
	if showOnlyNew {
		query = "SELECT user_id, name, url, description, address_street, address_locality, address_region, currency, price, is_new, uuid FROM listings WHERE user_id = ? AND is_new IS TRUE;"
	} else {
		query = "SELECT user_id, name, url, description, address_street, address_locality, address_region, currency, price, is_new, uuid FROM listings WHERE user_id = ?;"
	}

	result := make(listings.Listings, 0, defaultCapacity)
	rows, err := r.db.QueryContext(ctx, query, userID)
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
		if err = rows.Scan(&entry.UserID, &entry.Name, &entry.URL, &entry.Description, &entry.Address.StreetAddress, &entry.Address.AddressLocality, &entry.Address.AddressRegion, &entry.Offers.PriceCurrency, &entry.Offers.Price, &entry.IsNew, &entry.UUID); err != nil {
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

func (r *ListingsRepository) MGetListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (listings.Listings, error) {
	const name = "ListingsRepository.MGetListingByUserIDTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	result := make(listings.Listings, 0, defaultCapacity)
	rows, err := tx.QueryContext(ctx, "SELECT user_id, name, url, description, address_street, address_locality, address_region, currency, price, is_new, uuid FROM listings WHERE user_id = ?;", userID)
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
		if err = rows.Scan(&entry.UserID, &entry.Name, &entry.URL, &entry.Description, &entry.Address.StreetAddress, &entry.Address.AddressLocality, &entry.Address.AddressRegion, &entry.Offers.PriceCurrency, &entry.Offers.Price, &entry.IsNew, &entry.UUID); err != nil {
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

func (r *ListingsRepository) MInsertListingTx(ctx context.Context, tx domain.Tx, listings listings.Listings) error {
	if listings == nil || len(listings) == 0 {
		return nil
	}

	const fieldsLimit = 3275 // max is 32766 divided by 10
	if len(listings) <= fieldsLimit {
		return r.mInsertListingTx(ctx, tx, listings)
	}

	lbound := 0
	hbound := fieldsLimit
	for lbound < hbound {
		listingsSlice := listings[lbound:hbound]
		if err := r.mInsertListingTx(ctx, tx, listingsSlice); err != nil {
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

func (r *ListingsRepository) mInsertListingTx(ctx context.Context, tx domain.Tx, listings listings.Listings) error {
	const (
		name     = "ListingsRepository.mInsertListingTx"
		fieldsNb = 11
	)
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	// build query
	b := strings.Builder{}
	params := make([]interface{}, 0, len(listings)*fieldsNb)
	b.WriteString("INSERT INTO listings (user_id, name, url, description, address_street, address_locality, address_region, currency, price, is_new, uuid) VALUES ")
	counter := 0
	for idx := range listings {
		if counter > 0 {
			b.WriteString(", ")
		}
		b.WriteString("(" + strings.Repeat("?,", fieldsNb-1) + "?)")
		params = append(
			params,
			listings[idx].UserID,
			listings[idx].Name,
			listings[idx].URL,
			listings[idx].Description,
			listings[idx].Address.StreetAddress,
			listings[idx].Address.AddressLocality,
			listings[idx].Address.AddressRegion,
			listings[idx].Offers.PriceCurrency,
			listings[idx].Offers.Price,
			true,
			listings[idx].UUID,
		)
		counter++
	}

	_, err := tx.ExecContext(ctx, b.String(), params...)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *ListingsRepository) MUpdateListingTx(ctx context.Context, tx domain.Tx, listings listings.Listings) error {
	const name = "ListingsRepository.MUpdateListingTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	if listings == nil || len(listings) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, "UPDATE listings SET name = ?, description = ?, address_street = ?, address_locality = ?, address_region = ?, currency = ?, price = ?, is_new = false WHERE user_id = ? and url = ?;")
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to prepare statement in")
		return fmt.Errorf("failed to prepare statement in %s: %w", name, err)
	}
	defer stmt.Close()

	for idx := range listings {
		_, err = stmt.ExecContext(ctx, listings[idx].Name, listings[idx].Description, listings[idx].Address.StreetAddress, listings[idx].Address.AddressLocality, listings[idx].Address.AddressRegion, listings[idx].Offers.PriceCurrency, listings[idx].Offers.Price, listings[idx].UserID, listings[idx].URL)
		if err != nil {
			r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
			return fmt.Errorf("failed to execute query in %s: %w", name, err)
		}
	}

	return nil
}

func (r *ListingsRepository) UpdateFavoriteListingTx(ctx context.Context, tx domain.Tx, listing *listings.Listing) error {
	const name = "ListingsRepository.UpdateFavoriteListingTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	if listing == nil {
		return nil
	}

	_, err := tx.ExecContext(ctx, "UPDATE favorites SET name = ?, description = ?, address_street = ?, address_locality = ?, address_region = ?, currency = ?, price = ? WHERE user_id = ? and url = ?;", listing.Name, listing.Description, listing.Address.StreetAddress, listing.Address.AddressLocality, listing.Address.AddressRegion, listing.Offers.PriceCurrency, listing.Offers.Price, listing.UserID, listing.URL)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *ListingsRepository) InsertFavoriteListingTx(ctx context.Context, tx domain.Tx, listing *listings.Listing) error {
	const name = "ListingsRepository.InsertFavoriteListingTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	if listing == nil {
		return nil
	}

	_, err := tx.ExecContext(ctx, "INSERT INTO favorites (user_id, name, url, description, address_street, address_locality, address_region, currency, price) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);", listing.UserID, listing.Name, listing.URL, listing.Description, listing.Address.StreetAddress, listing.Address.AddressLocality, listing.Address.AddressRegion, listing.Offers.PriceCurrency, listing.Offers.Price)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *ListingsRepository) MGetFavoriteListingByUserID(ctx context.Context, userID string) (listings.Listings, error) {
	const name = "ListingsRepository.MGetFavoriteListingByUserID"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	result := make(listings.Listings, 0, defaultCapacity)
	rows, err := r.db.QueryContext(ctx, "SELECT user_id, name, url, description, address_street, address_locality, address_region, currency, price FROM favorites WHERE user_id = ?;", userID)
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
		if err = rows.Scan(&entry.UserID, &entry.Name, &entry.URL, &entry.Description, &entry.Address.StreetAddress, &entry.Address.AddressLocality, &entry.Address.AddressRegion, &entry.Offers.PriceCurrency, &entry.Offers.Price); err != nil {
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

func (r *ListingsRepository) MGetFavoriteListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (listings.Listings, error) {
	const name = "ListingsRepository.MGetFavoriteListingByUserIDTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	result := make(listings.Listings, 0, defaultCapacity)
	rows, err := tx.QueryContext(ctx, "SELECT user_id, name, url, description, address_street, address_locality, address_region, currency, price FROM favorites WHERE user_id = ?;", userID)
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
		if err = rows.Scan(&entry.UserID, &entry.Name, &entry.URL, &entry.Description, &entry.Address.StreetAddress, &entry.Address.AddressLocality, &entry.Address.AddressRegion, &entry.Offers.PriceCurrency, &entry.Offers.Price); err != nil {
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

func (r *ListingsRepository) MDeleteFavoriteListingByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error {
	const name = "ListingsRepository.MDeleteFavoriteListingByUserIDTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := tx.ExecContext(ctx, "DELETE FROM favorites WHERE user_id = ?;", userID)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}
