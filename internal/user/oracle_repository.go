package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"my-backlog/internal/apperrors"
	"my-backlog/internal/database"
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
	id, err := database.NewUUID()
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

// FindByEmail retorna o User com o e-mail informado.
// Retorna InputError{KindNotFound} se não existir.
func (r *OracleRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash FROM mbl_users WHERE email = :1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, &apperrors.InputError{
			Kind:    apperrors.KindNotFound,
			Field:   "email",
			Message: fmt.Sprintf("email %q not found", email),
		}
	}
	if err != nil {
		return User{}, &apperrors.InfraError{Op: fmt.Sprintf("find-by-email (email=%q)", email), Cause: err}
	}
	return u, nil
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
