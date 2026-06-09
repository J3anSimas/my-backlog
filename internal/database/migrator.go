package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
//	m := database.NewMigrator(db, migrations)
//	err := m.Up(ctx)
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator cria um Migrator com o banco e a lista de migrations injetados.
func NewMigrator(db *sql.DB, migrations []Migration) *Migrator {
	return &Migrator{db: db, migrations: migrations}
}

// LoadMigrations lê os arquivos SQL embarcados em migrations/*.sql e os retorna ordenados por versão.
func LoadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
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
		content, err := fs.ReadFile(migrationsFS, "migrations/"+entry.Name())
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
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return err
	}
	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return err
	}
	for _, mg := range m.pendingMigrations(applied) {
		if err := m.apply(ctx, mg); err != nil {
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

func (m *Migrator) ensureMigrationsTable(ctx context.Context) error {
	var count int
	err := m.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM user_tables WHERE table_name = 'SCHEMA_MIGRATIONS'",
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("checking schema_migrations table: %w", err)
	}
	if count > 0 {
		return nil
	}
	_, err = m.db.ExecContext(ctx, `
		CREATE TABLE schema_migrations (
			version    NUMBER        NOT NULL,
			name       VARCHAR2(255) NOT NULL,
			applied_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP NOT NULL,
			CONSTRAINT pk_schema_migrations PRIMARY KEY (version)
		)
	`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}
	return nil
}

func (m *Migrator) appliedVersions(ctx context.Context) (map[int]bool, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("querying applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scanning migration version: %w", err)
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// apply executa o SQL da migration e registra a versão aplicada.
// DDL Oracle realiza commit implícito, por isso o registro é feito logo após.
func (m *Migrator) apply(ctx context.Context, mg Migration) error {
	if _, err := m.db.ExecContext(ctx, mg.SQL); err != nil {
		return fmt.Errorf("executing sql: %w", err)
	}
	_, err := m.db.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, name) VALUES (:1, :2)",
		mg.Version, mg.Name,
	)
	return err
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
