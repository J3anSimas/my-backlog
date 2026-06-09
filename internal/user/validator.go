package user

import (
	"fmt"
	"strings"

	"my-backlog/internal/apperrors"
)

const (
	MaxNameLength  = 100
	MaxEmailLength = 254
	MinPasswordLen = 8
)

// Validator valida um User antes de persistir.
type Validator interface {
	Validate(u User) error
}

// StrictValidator implementa as regras de negócio padrão de cadastro.
type StrictValidator struct{}

// DefaultValidator retorna um Validator com as regras de negócio padrão.
func DefaultValidator() Validator { return StrictValidator{} }

// Validate retorna *apperrors.InputError se o User violar alguma constraint.
func (StrictValidator) Validate(u User) error {
	if strings.TrimSpace(u.Name) == "" {
		return &apperrors.InputError{
			Kind:    apperrors.KindInvalid,
			Field:   "name",
			Message: fmt.Sprintf("name must not be empty, got %q", u.Name),
		}
	}
	if len(u.Name) > MaxNameLength {
		return &apperrors.InputError{
			Kind:    apperrors.KindInvalid,
			Field:   "name",
			Message: fmt.Sprintf("name must be at most %d characters, got %d", MaxNameLength, len(u.Name)),
		}
	}
	if strings.TrimSpace(u.Email) == "" {
		return &apperrors.InputError{
			Kind:    apperrors.KindInvalid,
			Field:   "email",
			Message: fmt.Sprintf("email must not be empty, got %q", u.Email),
		}
	}
	if !strings.Contains(u.Email, "@") || len(u.Email) > MaxEmailLength {
		return &apperrors.InputError{
			Kind:    apperrors.KindInvalid,
			Field:   "email",
			Message: fmt.Sprintf("email must be a valid address, got %q", u.Email),
		}
	}
	if len(u.Password) < MinPasswordLen {
		return &apperrors.InputError{
			Kind:    apperrors.KindInvalid,
			Field:   "password",
			Message: fmt.Sprintf("password must be at least %d characters, got %d", MinPasswordLen, len(u.Password)),
		}
	}
	return nil
}
