package storage

import "context"

type MySQLRepository interface {
	Migrate(ctx context.Context, direction string) error
}
