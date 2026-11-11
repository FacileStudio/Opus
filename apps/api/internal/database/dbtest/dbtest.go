// Package dbtest gives a test a real PostgreSQL database carrying Opus's
// migrations. Each test binary owns one schema through tronc/testdb, and
// every Open empties it, so a test starts from no rows and fixed identity
// sequences. Without TEST_DATABASE_URL the test is skipped.
package dbtest

import (
	"context"
	"hash/fnv"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/database"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
)

const bootstrap = "docker run -d --name opus-test -e POSTGRES_PASSWORD=opus -p 127.0.0.1:5433:5432 postgres:16"

// Config is the testdb configuration every Opus test package shares.
func Config() testdb.Config {
	return testdb.Config{Prefix: "opus_test", Migrate: migrate}
}

// Open connects to this binary's schema, migrates it, empties every table and
// closes the pool when the test ends.
func Open(t testing.TB) *gorm.DB {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason(bootstrap))
	}
	db, err := testdb.Open(url, Config())
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql handle: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Logf("close test database: %v", err)
		}
	})
	if err := testdb.Truncate(db, Config()); err != nil {
		t.Fatalf("empty test database: %v", err)
	}
	return db
}

func migrate(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return database.RunMigrationsLocked(context.Background(), sqlDB, lockID(testdb.SchemaName(Config().Prefix)))
}

func lockID(schema string) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(schema))
	return int64(hash.Sum64() >> 1)
}
