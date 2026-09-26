package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/FacileStudio/Opus/apps/api/migrations"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// MigrationsFS returns the embedded filesystem containing SQL migrations.
func MigrationsFS() fs.FS {
	return migrations.FS
}

// RunMigrations applies all pending database migrations using goose.
func RunMigrations(db *sql.DB) error {
	return RunMigrationsContext(context.Background(), db)
}

// RunMigrationsContext applies all pending migrations with context and Postgres session lock.
func RunMigrationsContext(ctx context.Context, db *sql.DB) error {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("session locker: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS,
		goose.WithSessionLocker(locker),
		goose.WithDisableGlobalRegistry(true),
	)
	if err != nil {
		return fmt.Errorf("create goose provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
