package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const maintenanceDatabase = "postgres"

// EnsureDatabase creates the database named in databaseURL when it is missing,
// so a PostgreSQL volume that predates the init scripts does not stop the
// backend from booting. Keyword/value DSNs and URLs without a database name are
// left untouched, returning nil so Open reports the connection failure instead.
func EnsureDatabase(databaseURL string) error {
	adminURL, name, ok := splitDatabaseURL(databaseURL)
	if !ok {
		return nil
	}
	admin, err := gorm.Open(postgres.Open(adminURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("connect to %s: %w", maintenanceDatabase, err)
	}
	sqlDB, err := admin.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	row := sqlDB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", name)
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("check database %q: %w", name, err)
	}
	if exists {
		return nil
	}
	if _, err := sqlDB.ExecContext(ctx, "CREATE DATABASE "+quoteIdentifier(name)); err != nil {
		return fmt.Errorf("create database %q: %w", name, err)
	}
	return nil
}

// splitDatabaseURL separates a PostgreSQL connection URL into an admin URL that
// points at the maintenance database and the target database name. It reports
// ok=false for anything that is not a postgres:// or postgresql:// URL, or that
// does not name a database.
func splitDatabaseURL(databaseURL string) (adminURL string, name string, ok bool) {
	parsed, err := url.Parse(strings.TrimSpace(databaseURL))
	if err != nil {
		return "", "", false
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return "", "", false
	}
	name = strings.TrimPrefix(parsed.Path, "/")
	if name == "" {
		return "", "", false
	}
	admin := *parsed
	admin.Path = "/" + maintenanceDatabase
	return admin.String(), name, true
}

// quoteIdentifier wraps a database name in double quotes so names with mixed
// case or reserved words are accepted by CREATE DATABASE.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
