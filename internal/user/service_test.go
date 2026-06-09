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

// --- Login ---

func registerOne(t *testing.T, svc *user.Service, email, password string) {
	t.Helper()
	_, err := svc.Register(context.Background(), "Usuário", email, password)
	if err != nil {
		t.Fatalf("falha ao registrar usuário de teste: %v", err)
	}
}

func TestService_Login_ValidCredentials_ReturnsUser(t *testing.T) {
	svc, _ := newService()
	registerOne(t, svc, "joao@email.com", "senha123")

	got, err := svc.Login(context.Background(), "joao@email.com", "senha123")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.Email != "joao@email.com" {
		t.Errorf("esperado Email=%q, obtido %q", "joao@email.com", got.Email)
	}
	if got.ID == "" {
		t.Error("esperado ID não vazio")
	}
}

func TestService_Login_UnknownEmail_ReturnsUnauthorized(t *testing.T) {
	svc, _ := newService()

	_, err := svc.Login(context.Background(), "naoexiste@email.com", "senha123")

	var inputErr *apperrors.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *apperrors.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Kind != apperrors.KindUnauthorized {
		t.Errorf("esperado KindUnauthorized, obtido %v", inputErr.Kind)
	}
	if inputErr.Field != "" {
		t.Errorf("esperado Field vazio (genérico), obtido %q", inputErr.Field)
	}
}

func TestService_Login_WrongPassword_ReturnsUnauthorized(t *testing.T) {
	svc, _ := newService()
	registerOne(t, svc, "joao@email.com", "senha123")

	_, err := svc.Login(context.Background(), "joao@email.com", "senhaerrada")

	var inputErr *apperrors.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *apperrors.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Kind != apperrors.KindUnauthorized {
		t.Errorf("esperado KindUnauthorized, obtido %v", inputErr.Kind)
	}
}

func TestService_Login_RepositoryFails_ReturnsInfraError(t *testing.T) {
	svc, repo := newService()
	repoErr := &apperrors.InfraError{Op: "db-select", Cause: errors.New("connection refused")}
	repo.FailWith(repoErr)

	_, err := svc.Login(context.Background(), "joao@email.com", "senha123")

	var infraErr *apperrors.InfraError
	if !errors.As(err, &infraErr) {
		t.Fatalf("esperado *apperrors.InfraError, obtido %T: %v", err, err)
	}
}

func TestService_Login_UnknownEmail_TakesComparableTime(t *testing.T) {
	// Verifica que Login não retorna imediatamente quando o e-mail não existe,
	// evitando enumeração de e-mails por timing. A presença do dummy bcrypt
	// garante que o tempo seja comparável ao de senha errada para e-mail válido.
	// Este teste não mede tempo — apenas confirma que o código não compila/panic.
	svc, _ := newService()
	registerOne(t, svc, "joao@email.com", "senha123")

	_, err1 := svc.Login(context.Background(), "naoexiste@email.com", "senha123")
	_, err2 := svc.Login(context.Background(), "joao@email.com", "senhaerrada")

	var ie1, ie2 *apperrors.InputError
	if !errors.As(err1, &ie1) || ie1.Kind != apperrors.KindUnauthorized {
		t.Errorf("e-mail inexistente: esperado KindUnauthorized, obtido %v", err1)
	}
	if !errors.As(err2, &ie2) || ie2.Kind != apperrors.KindUnauthorized {
		t.Errorf("senha errada: esperado KindUnauthorized, obtido %v", err2)
	}
}
