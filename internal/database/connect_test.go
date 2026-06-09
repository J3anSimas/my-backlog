package database_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"my-backlog/internal/database"
)

func init() {
	sql.Register("fakedb", fakeDriver{})
}

// fakeDriver implementa um driver SQL mínimo para testes unitários sem Oracle.
// Retorna count=0 para queries em user_tables, rows vazias para demais selects.

type fakeDriver struct{}

func (fakeDriver) Open(name string) (driver.Conn, error) { return &fakeConn{}, nil }

type fakeConn struct{}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{query: query}, nil
}
func (c *fakeConn) Close() error { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) {
	return nil, errors.New("fakedb: transactions not supported")
}

type fakeStmt struct{ query string }

func (s *fakeStmt) Close() error  { return nil }
func (s *fakeStmt) NumInput() int { return -1 }

func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) { return fakeResult{}, nil }

func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	if strings.Contains(strings.ToLower(s.query), "user_tables") {
		return &singleInt64Row{value: 0}, nil
	}
	return &emptyRows{}, nil
}

type fakeResult struct{}

func (fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (fakeResult) RowsAffected() (int64, error) { return 1, nil }

type singleInt64Row struct {
	value int64
	done  bool
}

func (r *singleInt64Row) Columns() []string { return []string{"count"} }
func (r *singleInt64Row) Close() error      { return nil }
func (r *singleInt64Row) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	dest[0] = r.value
	r.done = true
	return nil
}

type emptyRows struct{}

func (r *emptyRows) Columns() []string              { return nil }
func (r *emptyRows) Close() error                   { return nil }
func (r *emptyRows) Next(dest []driver.Value) error { return io.EOF }

// testOpener abre uma conexão fake via fakedb, sem Oracle real.
type testOpener struct{}

func (testOpener) Open(_ context.Context, _ database.Config) (*sql.DB, error) {
	return sql.Open("fakedb", "test")
}

// --- Ciclo 1: tracer bullet ---

func TestConnect_WithFakeOpener_ReturnsDB(t *testing.T) {
	db, err := database.Connect(database.WithOpener(testOpener{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil *sql.DB")
	}
	db.Close()
}

// --- Ciclo 2: ErrNotConfigured sem env vars ---

func TestConnect_SemEnvVars_RetornaErrNotConfigured(t *testing.T) {
	t.Setenv("DB_USER", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_SERVICE_NAME", "")

	_, err := database.Connect()
	if !errors.Is(err, database.ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}

// --- Ciclo 3: config explicitamente vazia sem opener ---

func TestConnect_ConfigVazia_SemOpener_RetornaErrNotConfigured(t *testing.T) {
	_, err := database.Connect(database.WithConfig(database.Config{}))
	if !errors.Is(err, database.ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}

// --- Ciclo 4: WithMigrations chama a fonte ---

type countingSource struct {
	calls      int
	migrations []database.Migration
}

func (s *countingSource) Migrations() ([]database.Migration, error) {
	s.calls++
	return s.migrations, nil
}

func TestConnect_ComMigrations_ChamamFonteDeMigrations(t *testing.T) {
	source := &countingSource{}

	db, err := database.Connect(
		database.WithOpener(testOpener{}),
		database.WithMigrations(source),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	db.Close()

	if source.calls == 0 {
		t.Error("expected migration source to be called at least once")
	}
}

// --- Ciclo 5: falha na fonte de migrations retorna erro ---

type failingSource struct{}

func (failingSource) Migrations() ([]database.Migration, error) {
	return nil, errors.New("migration load failure")
}

func TestConnect_FalhaNaFonteDeMigrations_RetornaErro(t *testing.T) {
	_, err := database.Connect(
		database.WithOpener(testOpener{}),
		database.WithMigrations(failingSource{}),
	)
	if err == nil {
		t.Fatal("expected error from failing migration source, got nil")
	}
}

// --- Ciclo 6: OpenForApp sem env vars ---

func TestOpenForApp_SemEnvVars_RetornaErrNotConfigured(t *testing.T) {
	t.Setenv("DB_USER", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_SERVICE_NAME", "")

	_, err := database.OpenForApp()
	if !errors.Is(err, database.ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}
