package database

import (
	"context"
	"database/sql"
	"fmt"
)

// OracleMigrationStore implementa MigrationStore usando *sql.DB com Oracle.
// Apply inverte a ordem tradicional: INSERT antes do DDL, eliminando o split-brain
// onde o DDL commitava implicitamente antes do registro ser persistido.
type OracleMigrationStore struct {
	db *sql.DB
}

// NewOracleMigrationStore cria um OracleMigrationStore para o banco fornecido.
func NewOracleMigrationStore(db *sql.DB) *OracleMigrationStore {
	return &OracleMigrationStore{db: db}
}

// EnsureTracking cria a tabela mbl_schema_migrations se ainda não existir.
func (s *OracleMigrationStore) EnsureTracking(ctx context.Context) error {
	var count int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM user_tables WHERE table_name = 'MBL_SCHEMA_MIGRATIONS'",
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("checking mbl_schema_migrations table: %w", err)
	}
	if count > 0 {
		return nil
	}
	_, err = s.db.ExecContext(ctx, `
		CREATE TABLE mbl_schema_migrations (
			version    NUMBER        NOT NULL,
			name       VARCHAR2(255) NOT NULL,
			applied_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP NOT NULL,
			CONSTRAINT pk_mbl_schema_migrations PRIMARY KEY (version)
		)
	`)
	if err != nil {
		return fmt.Errorf("creating mbl_schema_migrations table: %w", err)
	}
	return nil
}

// AppliedVersions retorna o conjunto de versões já registradas em mbl_schema_migrations.
func (s *OracleMigrationStore) AppliedVersions(ctx context.Context) (map[int]bool, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT version FROM mbl_schema_migrations")
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

// Apply registra a versão antes de executar o DDL.
// Se o DDL falhar, tenta desfazer o registro. Se esse rollback também falhar,
// retorna ErrDDLAppliedUntracked — estado irrecuperável que exige intervenção manual.
func (s *OracleMigrationStore) Apply(ctx context.Context, m Migration) error {
	if err := s.insertRecord(ctx, m); err != nil {
		return fmt.Errorf("recording migration %d before DDL: %w", m.Version, err)
	}
	if _, err := s.db.ExecContext(ctx, m.SQL); err != nil {
		if rollbackErr := s.deleteRecord(ctx, m.Version); rollbackErr != nil {
			return fmt.Errorf("DDL for migration %d failed and record rollback also failed: %w: %v", m.Version, ErrDDLAppliedUntracked, rollbackErr)
		}
		return fmt.Errorf("executing DDL %d: %w", m.Version, err)
	}
	return nil
}

func (s *OracleMigrationStore) insertRecord(ctx context.Context, m Migration) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO mbl_schema_migrations (version, name) VALUES (:1, :2)",
		m.Version, m.Name,
	)
	return err
}

func (s *OracleMigrationStore) deleteRecord(ctx context.Context, version int) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM mbl_schema_migrations WHERE version = :1",
		version,
	)
	return err
}
