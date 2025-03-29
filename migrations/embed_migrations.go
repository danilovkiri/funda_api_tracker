package tsdbmigrations

import "embed"

//go:embed *.sql

// EmbedMigrations is used in migration
var EmbedMigrations embed.FS
