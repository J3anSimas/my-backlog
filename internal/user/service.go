package user

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"my-backlog/internal/apperrors"
)

// dummyHash é usado quando o e-mail não existe, para evitar enumeração por timing:
// bcrypt.CompareHashAndPassword sempre roda, independente de o usuário existir.
var dummyHash = []byte("$2a$12$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

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

// Login autentica um usuário pelo e-mail e senha.
// Retorna InputError{KindUnauthorized} para qualquer falha de credencial,
// sem indicar qual campo está errado.
func (s *Service) Login(ctx context.Context, email, password string) (User, error) {
	u, err := s.repo.FindByEmail(ctx, email)

	var notFound *apperrors.InputError
	if err != nil {
		if !errors.As(err, &notFound) {
			return User{}, &apperrors.InfraError{Op: "find-by-email", Cause: err}
		}
		// e-mail não existe: ainda compara com dummy para equalizar tempo
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password)) //nolint:errcheck
		return User{}, &apperrors.InputError{Kind: apperrors.KindUnauthorized, Message: "credenciais inválidas"}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return User{}, &apperrors.InputError{Kind: apperrors.KindUnauthorized, Message: "credenciais inválidas"}
	}
	return u, nil
}
