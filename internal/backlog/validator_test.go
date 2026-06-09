package backlog_test

import (
	"errors"
	"strings"
	"testing"

	"my-backlog/internal/backlog"
)

func TestStrictValidator_EmptyTitle_ReturnsInputError(t *testing.T) {
	v := backlog.DefaultValidator()

	err := v.Validate(backlog.Backlog{Title: "", Description: "válida"})

	var inputErr *backlog.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *backlog.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "title" {
		t.Errorf("esperado Field=%q, obtido %q", "title", inputErr.Field)
	}
	if inputErr.Kind != backlog.KindInvalid {
		t.Errorf("esperado Kind=KindInvalid, obtido %v", inputErr.Kind)
	}
}

func TestStrictValidator_WhitespaceOnlyTitle_ReturnsInputError(t *testing.T) {
	v := backlog.DefaultValidator()

	err := v.Validate(backlog.Backlog{Title: "   ", Description: "válida"})

	var inputErr *backlog.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *backlog.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "title" {
		t.Errorf("esperado Field=%q, obtido %q", "title", inputErr.Field)
	}
}

func TestStrictValidator_TitleExceedsMax_ReturnsInputError(t *testing.T) {
	v := backlog.DefaultValidator()
	longTitle := strings.Repeat("a", backlog.MaxTitleLength+1)

	err := v.Validate(backlog.Backlog{Title: longTitle, Description: "válida"})

	var inputErr *backlog.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *backlog.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "title" {
		t.Errorf("esperado Field=%q, obtido %q", "title", inputErr.Field)
	}
}

func TestStrictValidator_EmptyDescription_ReturnsInputError(t *testing.T) {
	v := backlog.DefaultValidator()

	err := v.Validate(backlog.Backlog{Title: "válido", Description: ""})

	var inputErr *backlog.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *backlog.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "description" {
		t.Errorf("esperado Field=%q, obtido %q", "description", inputErr.Field)
	}
}

func TestStrictValidator_DescriptionExceedsMax_ReturnsInputError(t *testing.T) {
	v := backlog.DefaultValidator()
	longDesc := strings.Repeat("a", backlog.MaxDescriptionLength+1)

	err := v.Validate(backlog.Backlog{Title: "válido", Description: longDesc})

	var inputErr *backlog.InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("esperado *backlog.InputError, obtido %T: %v", err, err)
	}
	if inputErr.Field != "description" {
		t.Errorf("esperado Field=%q, obtido %q", "description", inputErr.Field)
	}
}

func TestStrictValidator_ValidBacklog_ReturnsNil(t *testing.T) {
	v := backlog.DefaultValidator()

	err := v.Validate(backlog.Backlog{Title: "Título válido", Description: "Descrição válida"})

	if err != nil {
		t.Errorf("esperado nil para backlog válido, obtido %v", err)
	}
}
