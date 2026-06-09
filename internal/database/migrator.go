package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// ErrDDLAppliedUntracked is returned by MigrationStore.Apply when DDL committed
// successfully but the version record could not be persisted, leaving the schema
// in an inconsistent state that requires manual intervention.
var ErrDDLAppliedUntracked = errors.New("DDL committed but version record failed")

// MigrationStore abstracts the persistence layer used by Migrator.
// Implement this interface to swap Oracle for a fake in unit tests.
//
// Example fake:
//
//	type noopStore struct{}
//	func (noopStore) EnsureTracking(_ context.Context) error              { return nil }
//	func (noopStore) AppliedVersions(_ context.Context) (map[int]bool, error) { return map[int]bool{}, nil }
//	func (noopStore) Apply(_ context.Context, _ Migration) error          { return nil }
type MigrationStore interface {
	EnsureTracking(ctx context.Context) error
	AppliedVersions(ctx context.Context) (map[int]bool, error)
	// Apply records and executes a single migration.
	// Must return ErrDDLAppliedUntracked when DDL committed but the version record failed.
	Apply(ctx context.Context, m Migration) error
}

// Migration representa um arquivo SQL versionado a ser aplicado uma única vez.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Migrator aplica migrations pendentes em ordem crescente de versão.
//
// Exemplo:
//
//	migrations, _ := database.LoadMigrations()
//	m := database.NewMigrator(database.NewOracleMigrationStore(db), migrations)
//	err := m.Up(ctx)
type Migrator struct {
	store      MigrationStore
	migrations []Migration
}

// NewMigrator cria um Migrator com o store e a lista de migrations injetados.
func NewMigrator(store MigrationStore, migrations []Migration) *Migrator {
	return &Migrator{store: store, migrations: migrations}
}

// LoadMigrations lê os arquivos SQL embarcados em migrations/*.sql e os retorna ordenados por versão.
func LoadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(MigrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("reading embedded migrations: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, name, err := parseMigrationFilename(entry.Name())
		if err != nil {
			return nil, err
		}
		content, err := fs.ReadFile(MigrationsFS, "migrations/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading migration file %s: %w", entry.Name(), err)
		}
		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     normalizeMigrationSQL(string(content)),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

// Up cria a tabela de controle se necessário e aplica todas as migrations pendentes.
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.store.EnsureTracking(ctx); err != nil {
		return err
	}
	applied, err := m.store.AppliedVersions(ctx)
	if err != nil {
		return err
	}
	for _, mg := range m.pendingMigrations(applied) {
		if err := m.store.Apply(ctx, mg); err != nil {
			return fmt.Errorf("applying migration %d (%s): %w", mg.Version, mg.Name, err)
		}
	}
	return nil
}

// pendingMigrations retorna as migrations ainda não aplicadas, em ordem.
func (m *Migrator) pendingMigrations(applied map[int]bool) []Migration {
	var pending []Migration
	for _, mg := range m.migrations {
		if !applied[mg.Version] {
			pending = append(pending, mg)
		}
	}
	return pending
}

// parseMigrationFilename extrai versão e nome de um arquivo no formato NNNN_nome.sql.
func parseMigrationFilename(filename string) (version int, name string, err error) {
	base := strings.TrimSuffix(filename, ".sql")
	idx := strings.Index(base, "_")
	if idx < 0 {
		return 0, "", fmt.Errorf("invalid migration filename %q: expected NNNN_name.sql format", filename)
	}
	v, err := strconv.Atoi(base[:idx])
	if err != nil {
		return 0, "", fmt.Errorf("invalid version prefix in migration filename %q: expected integer, got %q", filename, base[:idx])
	}
	return v, base[idx+1:], nil
}

// normalizeMigrationSQL remove espaços e ponto-e-vírgula final que o driver Oracle não aceita em DDL.
func normalizeMigrationSQL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimRight(s, ";")
	return strings.TrimSpace(s)
}
