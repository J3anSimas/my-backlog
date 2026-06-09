package user_test

import (
	"context"
	"errors"
	"testing"

	"my-backlog/internal/apperrors"
	"my-backlog/internal/user"
	"my-backlog/internal/user/usertest"
)

func newService() (*user.Service, *usertest.FakeRepository) {
	repo := usertest.NewFakeRepository()
	return user.NewService(repo, user.DefaultValidator()), repo
}

func TestService_Register_ReturnsUserWithIDAndHashedPassword(t *testing.T) {
	svc, _ := newService()

	got, err := svc.Register(context.Background(), "João", "joao@email.com", "senha123")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.ID == "" {
		t.Error("esperado ID não vazio")
	}
	if got.Email != "joao@email.com" {
		t.Errorf("esperado Email=%q, obtido %q", "joao@email.com", got.Email)
	}
	if got.PasswordHash == "" {
		t.Error("esperado PasswordHash não vazio")
	}
	if got.PasswordHash == "senha123" {
		t.Error("PasswordHash não deve ser o texto plano da senha")
	}
}

func TestService_Register_PersistsUserInRepository(t *testing.T) {
	svc, repo := newService()

	_, err := svc.Register(context.Background(), "João", "joao@email.com", "senha123")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.SaveCount() != 1 {
		t.Errorf("esperado 1 usuário salvo, obtido %d", repo.SaveCount())
	}
}

func TestService_Register_InvalidInput_ReturnsInputError(t *testing.T) {
	svc, _ := newService()

	_, err := svc.Register(context.Background(), "", "joao@email.com", "senha123")

	var inputErr *apperrors.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *apperrors.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "name" {
		t.Errorf("esperado Field=%q, obtido %q", "name", inputErr.Field)
	}
}

func TestService_Register_DuplicateEmail_ReturnsConflictError(t *testing.T) {
	svc, _ := newService()

	_, err := svc.Register(context.Background(), "João", "joao@email.com", "senha123")
	if err != nil {
		t.Fatalf("erro no primeiro cadastro: %v", err)
	}

	_, err = svc.Register(context.Background(), "João 2", "joao@email.com", "senha456")

	var inputErr *apperrors.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *apperrors.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Kind != apperrors.KindConflict {
		t.Errorf("esperado KindConflict, obtido %v", inputErr.Kind)
	}
	if inputErr.Field != "email" {
		t.Errorf("esperado Field=%q, obtido %q", "email", inputErr.Field)
	}
}

func TestService_Register_RepositoryFails_PropagatesInfraError(t *testing.T) {
	svc, repo := newService()
	repoErr := &apperrors.InfraError{Op: "db-insert", Cause: errors.New("connection refused")}
	repo.FailWith(repoErr)

	_, err := svc.Register(context.Background(), "João", "joao@email.com", "senha123")

	var infraErr *apperrors.InfraError
	if !errors.As(err, &infraErr) {
		t.Fatalf("esperado *apperrors.InfraError, obtido %T: %v", err, err)
	}
}
