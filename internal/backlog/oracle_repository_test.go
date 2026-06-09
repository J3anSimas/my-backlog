//go:build integration

package backlog_test

import (
	"context"
	"database/sql"
	"testing"

	"my-backlog/internal/backlog"
	"my-backlog/internal/database"
)

// Para rodar: go test -tags integration ./internal/backlog/...
// Variáveis necessárias: DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_SERVICE_NAME, DB_WALLET_PATH

func TestOracleRepository_Save_PersistsBacklog(t *testing.T) {
	db := openTestDB(t)
	runMigrations(t, db)

	repo := backlog.NewOracleRepository(db)
	svc := backlog.NewService(repo)

	got, err := svc.Create("Integration Backlog", "Criado via teste de integração")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID")
	}
	if got.Title != "Integration Backlog" {
		t.Errorf("expected title %q, got %q", "Integration Backlog", got.Title)
	}
	if got.Description != "Criado via teste de integração" {
		t.Errorf("expected description %q, got %q", "Criado via teste de integração", got.Description)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM backlogs WHERE id = :1", got.ID)
	})
}

func TestOracleRepository_Save_GeneratesUniqueIDs(t *testing.T) {
	db := openTestDB(t)
	runMigrations(t, db)

	repo := backlog.NewOracleRepository(db)
	svc := backlog.NewService(repo)

	a, errA := svc.Create("Backlog A", "Descrição A")
	if errA != nil {
		t.Fatalf("unexpected error on first create: %v", errA)
	}
	b, errB := svc.Create("Backlog B", "Descrição B")
	if errB != nil {
		t.Fatalf("unexpected error on second create: %v", errB)
	}
	if a.ID == b.ID {
		t.Errorf("expected unique IDs, got same ID %q for both", a.ID)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM backlogs WHERE id IN (:1, :2)", a.ID, b.ID)
	})
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	cfg := database.ConfigFromEnv()
	if cfg.User == "" {
		t.Skip("DB_USER not set — skipping integration test")
	}
	db, err := database.NewOracleDB(cfg)
	if err != nil {
		t.Fatalf("connecting to oracle: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func runMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	migrations, err := database.LoadMigrations()
	if err != nil {
		t.Fatalf("loading migrations: %v", err)
	}
	m := database.NewMigrator(db, migrations)
	if err := m.Up(context.Background()); err != nil {
		t.Fatalf("running migrations: %v", err)
	}
}
