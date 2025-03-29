package domain

import (
	"context"
	"database/sql"
)

type Tx interface {
	Commit() error
	Rollback() error
	QueryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, sql string, args ...any) *sql.Row
	ExecContext(ctx context.Context, sql string, arguments ...any) (sql.Result, error)
}
