package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
)

type stubConnector struct{}

func (stubConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, errors.New("stub: no database")
}

func (stubConnector) Driver() driver.Driver { return stubDriver{} }

type stubDriver struct{}

func (stubDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("stub: no database")
}

func TestMigrationsFS(t *testing.T) {
	f, err := MigrationsFS().Open("00001_baseline_opus.sql")
	if err != nil {
		t.Fatalf("failed to open baseline migration: %v", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read baseline migration: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("baseline migration is empty")
	}
}

func TestRunMigrationsWithStubDB(t *testing.T) {
	db := sql.OpenDB(stubConnector{})
	defer db.Close()

	err := RunMigrations(db)
	if err == nil {
		t.Fatal("expected error with stub database, got nil")
	}
}
