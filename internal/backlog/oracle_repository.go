package backlog

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
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
	id, err := newUUID()
	if err != nil {
		return Backlog{}, &InfraError{Op: "generate-id", Cause: err}
	}
	b.ID = id

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO backlogs (id, title, description) VALUES (:1, :2, :3)",
		b.ID, b.Title, b.Description,
	)
	if err != nil {
		return Backlog{}, &InfraError{Op: fmt.Sprintf("db-insert (title=%q)", b.Title), Cause: err}
	}
	return b, nil
}

// newUUID gera um UUID v4 usando crypto/rand sem dependências externas.
func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
