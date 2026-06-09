package backlog

import (
	"context"
	"database/sql"
	"fmt"

	"my-backlog/internal/database"
)

// OracleRepository implementa Repository usando Oracle via database/sql.
//
// Exemplo:
//
//	repo := backlog.NewOracleRepository(db)
//	svc := backlog.NewService(repo)
type OracleRepository struct {
	db *sql.DB
}

// NewOracleRepository cria um OracleRepository com o *sql.DB injetado.
func NewOracleRepository(db *sql.DB) *OracleRepository {
	return &OracleRepository{db: db}
}

// Save persiste um Backlog no Oracle e retorna o registro com ID preenchido.
func (r *OracleRepository) Save(ctx context.Context, b Backlog) (Backlog, error) {
	id, err := database.NewUUID()
	if err != nil {
		return Backlog{}, &InfraError{Op: "generate-id", Cause: err}
	}
	b.ID = id

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO mbl_backlogs (id, title, description) VALUES (:1, :2, :3)",
		b.ID, b.Title, b.Description,
	)
	if err != nil {
		return Backlog{}, &InfraError{Op: fmt.Sprintf("db-insert (title=%q)", b.Title), Cause: err}
	}
	return b, nil
}

