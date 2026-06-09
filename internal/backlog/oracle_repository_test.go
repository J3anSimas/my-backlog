//go:build integration

package backlog_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"my-backlog/internal/backlog"
	"my-backlog/internal/database"
)

// Para rodar: go test -tags integration ./internal/backlog/...
// Variáveis necessárias: DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_SERVICE_NAME, DB_WALLET_PATH

func TestOracleRepository_Save_PersistsBacklog(t *testing.T) {
	db := openTestDB(t)

	repo := backlog.NewOracleRepository(db)
	svc := backlog.NewService(repo)

	got, err := svc.Create(context.Background(), "Integration Backlog", "Criado via teste de integração")
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
		db.ExecContext(context.Background(), "DELETE FROM mbl_backlogs WHERE id = :1", got.ID)
	})
}

func TestOracleRepository_Save_GeneratesUniqueIDs(t *testing.T) {
	db := openTestDB(t)

	repo := backlog.NewOracleRepository(db)
	svc := backlog.NewService(repo)

	a, errA := svc.Create(context.Background(), "Backlog A", "Descrição A")
	if errA != nil {
		t.Fatalf("unexpected error on first create: %v", errA)
	}
	b, errB := svc.Create(context.Background(), "Backlog B", "Descrição B")
	if errB != nil {
		t.Fatalf("unexpected error on second create: %v", errB)
	}
	if a.ID == b.ID {
		t.Errorf("expected unique IDs, got same ID %q for both", a.ID)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM mbl_backlogs WHERE id IN (:1, :2)", a.ID, b.ID)
	})
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Connect(
		database.WithConfig(database.ConfigFromEnv()),
		database.WithMigrations(database.MigrationSourceFunc(database.LoadMigrations)),
	)
	if errors.Is(err, database.ErrNotConfigured) {
		t.Skip("Oracle env vars não configurados — pulando teste de integração")
	}
	if err != nil {
		t.Fatalf("connecting to oracle: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
