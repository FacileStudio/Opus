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
	return RunMigrationsLocked(ctx, db, lock.DefaultLockID)
}

// RunMigrationsLocked applies all pending migrations under the Postgres
// session lock lockID. The lock is database-wide, so callers sharing one
// database across schemas, as the tests do, pass one id per schema.
func RunMigrationsLocked(ctx context.Context, db *sql.DB, lockID int64) error {
	locker, err := lock.NewPostgresSessionLocker(lock.WithLockID(lockID))
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
