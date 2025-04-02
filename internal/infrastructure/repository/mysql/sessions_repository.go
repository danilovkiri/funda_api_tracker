package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"fundaNotifier/internal/domain"
	"fundaNotifier/internal/domain/sessions"
	"time"
)

var _ sessions.Repository = (*SessionsRepository)(nil)

type SessionsRepository struct {
	*Repository
}

func NewSessionsRepository(repository *Repository) *SessionsRepository {
	return &SessionsRepository{
		Repository: repository,
	}
}

func (r *SessionsRepository) CreateDefaultSession(ctx context.Context, userID string, chatID int64) error {
	const name = "SessionsRepository.CreateDefaultSession"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "INSERT INTO sessions (user_id, chat_id) VALUES (?, ?);", userID, chatID)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *SessionsRepository) SessionExistsByUserID(ctx context.Context, userID string) (bool, error) {
	const name = "SessionsRepository.ExistsByUserID"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT user_id FROM sessions WHERE user_id = ?);", userID).Scan(&exists)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return exists, fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return exists, nil
}

func (r *SessionsRepository) GetSessionByUserIDTx(ctx context.Context, tx domain.Tx, userID string) (*sessions.Session, error) {
	const name = "SessionsRepository.GetSessionByUserIDTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	var session sessions.Session
	err := tx.QueryRowContext(ctx, "SELECT user_id, chat_id, update_interval_seconds, is_active, regions, cities, last_synced_at FROM sessions WHERE user_id = ?;", userID).Scan(&session.UserID, &session.ChatID, &session.UpdateIntervalSeconds, &session.IsActive, &session.Regions, &session.Cities, &session.LastSyncedAt)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return nil, fmt.Errorf("failed to execute query in %s: %w", name, err)
	}
	session.ParseRawRegionsAndCities()

	return &session, nil
}

func (r *SessionsRepository) GetSessionByUserID(ctx context.Context, userID string) (*sessions.Session, error) {
	const name = "SessionsRepository.GetSessionByUserID"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	var session sessions.Session
	err := r.db.QueryRowContext(ctx, "SELECT user_id, chat_id, update_interval_seconds, is_active, regions, cities, last_synced_at FROM sessions WHERE user_id = ?;", userID).Scan(&session.UserID, &session.ChatID, &session.UpdateIntervalSeconds, &session.IsActive, &session.Regions, &session.Cities, &session.LastSyncedAt)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return nil, fmt.Errorf("failed to execute query in %s: %w", name, err)
	}
	session.ParseRawRegionsAndCities()

	return &session, nil
}

func (r *SessionsRepository) DeleteSessionByUserIDTx(ctx context.Context, tx domain.Tx, userID string) error {
	const name = "SessionsRepository.DeleteSessionByUserIDTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = ?;", userID)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *SessionsRepository) UpdateSessionByUserIDTx(ctx context.Context, tx domain.Tx, session *sessions.Session) error {
	const name = "SessionsRepository.UpsertSessionByUserIDTx"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	_, err := tx.ExecContext(ctx, "UPDATE sessions SET update_interval_seconds = ?, is_active = ?, regions = ?, cities = ?, last_synced_at = ? WHERE user_id = ?;", session.UpdateIntervalSeconds, session.IsActive, session.RegionsRaw, session.CitiesRaw, session.LastSyncedAt, session.UserID)
	if err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return fmt.Errorf("failed to execute query in %s: %w", name, err)
	}

	return nil
}

func (r *SessionsRepository) GetActiveSessions(ctx context.Context) (sessions.Sessions, error) {
	const name = "SessionsRepository.GetActiveSessions"
	ctx, cancel := context.WithTimeout(ctx, time.Second*defaultTimeoutSeconds)
	defer cancel()

	result := make(sessions.Sessions, 0, defaultCapacity)
	rows, err := r.db.QueryContext(ctx, "SELECT user_id, chat_id, update_interval_seconds, is_active, regions, cities, last_synced_at FROM sessions WHERE is_active IS TRUE")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		r.log.Error().Err(err).Str("method", name).Msg("failed to execute query in")
		return nil, fmt.Errorf("failed to execute query in %s: %w", name, err)
	}
	defer rows.Close()

	// iterate over rows
	for rows.Next() {
		var session sessions.Session
		if err = rows.Scan(&session.UserID, &session.ChatID, &session.UpdateIntervalSeconds, &session.IsActive, &session.Regions, &session.Cities, &session.LastSyncedAt); err != nil {
			r.log.Error().Err(err).Str("method", name).Msg("failed to scan a row in")
			return nil, fmt.Errorf("failed to scan a row in %s: %w", name, err)
		}
		session.ParseRawRegionsAndCities()
		result = append(result, session)
	}
	if err = rows.Err(); err != nil {
		r.log.Error().Err(err).Str("method", name).Msg("failed to iterate over rows in")
		return nil, fmt.Errorf("failed to iterate over rows in %s: %w", name, err)
	}

	return result, nil
}
