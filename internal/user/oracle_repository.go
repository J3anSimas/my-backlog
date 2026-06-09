package user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"

	"my-backlog/internal/apperrors"
)

// OracleRepository implementa Repository usando Oracle via database/sql.
//
// Exemplo:
//
//	repo := user.NewOracleRepository(db)
//	svc := user.NewService(repo, user.DefaultValidator())
type OracleRepository struct {
	db *sql.DB
}

// NewOracleRepository cria um OracleRepository com o *sql.DB injetado.
func NewOracleRepository(db *sql.DB) *OracleRepository {
	return &OracleRepository{db: db}
}

// Save persiste um User no Oracle e retorna o registro com ID preenchido.
func (r *OracleRepository) Save(ctx context.Context, u User) (User, error) {
	id, err := newUUID()
	if err != nil {
		return User{}, &apperrors.InfraError{Op: "generate-id", Cause: err}
	}
	u.ID = id

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO mbl_users (id, name, email, password_hash) VALUES (:1, :2, :3, :4)`,
		u.ID, u.Name, u.Email, u.PasswordHash,
	)
	if err != nil {
		return User{}, &apperrors.InfraError{Op: fmt.Sprintf("db-insert (email=%q)", u.Email), Cause: err}
	}
	return u, nil
}

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// EmailExists retorna true se já existe um usuário com o e-mail informado.
func (r *OracleRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM mbl_users WHERE email = :1`, email,
	).Scan(&count)
	if err != nil {
		return false, &apperrors.InfraError{Op: fmt.Sprintf("email-exists (email=%q)", email), Cause: err}
	}
	return count > 0, nil
}
