package user

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"my-backlog/internal/apperrors"
)

// Service executa as regras de negócio de usuários.
//
// Exemplo:
//
//	svc := user.NewService(repo, user.DefaultValidator())
//	u, err := svc.Register(ctx, "João", "joao@email.com", "senha123")
type Service struct {
	repo      Repository
	validator Validator
}

// NewService cria um Service com o repositório e validador injetados.
func NewService(repo Repository, v Validator) *Service {
	return &Service{repo: repo, validator: v}
}

// Register valida, faz hash da senha e persiste um novo User.
func (s *Service) Register(ctx context.Context, name, email, password string) (User, error) {
	u := User{Name: name, Email: email, Password: password}
	if err := s.validator.Validate(u); err != nil {
		return User{}, err
	}

	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return User{}, &apperrors.InfraError{Op: "email-exists", Cause: err}
	}
	if exists {
		return User{}, &apperrors.InputError{
			Kind:    apperrors.KindConflict,
			Field:   "email",
			Message: fmt.Sprintf("email %q already registered", email),
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, &apperrors.InfraError{Op: "bcrypt", Cause: err}
	}
	u.PasswordHash = string(hash)
	u.Password = "" // nunca persistir texto plano

	return s.repo.Save(ctx, u)
}
