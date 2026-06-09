package user_test

import (
	"errors"
	"testing"

	"my-backlog/internal/apperrors"
	"my-backlog/internal/user"
)

func TestStrictUserValidator_ValidUser_ReturnsNil(t *testing.T) {
	v := user.DefaultValidator()

	err := v.Validate(user.User{Name: "João", Email: "joao@email.com", Password: "senha123"})

	if err != nil {
		t.Errorf("esperado nil para usuário válido, obtido %v", err)
	}
}

func TestStrictUserValidator_EmptyName_ReturnsInputError(t *testing.T) {
	v := user.DefaultValidator()

	err := v.Validate(user.User{Name: "", Email: "a@b.com", Password: "senha123"})

	var inputErr *apperrors.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *apperrors.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "name" {
		t.Errorf("esperado Field=%q, obtido %q", "name", inputErr.Field)
	}
	if inputErr.Kind != apperrors.KindInvalid {
		t.Errorf("esperado KindInvalid, obtido %v", inputErr.Kind)
	}
}

func TestStrictUserValidator_InvalidEmail_ReturnsInputError(t *testing.T) {
	v := user.DefaultValidator()

	err := v.Validate(user.User{Name: "João", Email: "nao-e-email", Password: "senha123"})

	var inputErr *apperrors.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *apperrors.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "email" {
		t.Errorf("esperado Field=%q, obtido %q", "email", inputErr.Field)
	}
}

func TestStrictUserValidator_ShortPassword_ReturnsInputError(t *testing.T) {
	v := user.DefaultValidator()

	err := v.Validate(user.User{Name: "João", Email: "joao@email.com", Password: "1234567"})

	var inputErr *apperrors.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *apperrors.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "password" {
		t.Errorf("esperado Field=%q, obtido %q", "password", inputErr.Field)
	}
}
