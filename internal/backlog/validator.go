package backlog

import (
	"fmt"
	"strings"
)

// MaxTitleLength e MaxDescriptionLength espelham VARCHAR2(N) em 0001_create_backlogs.sql.
// Alterar aqui exige alterar a migration correspondente.
const (
	MaxTitleLength       = 200
	MaxDescriptionLength = 1000
)

// Validator valida um Backlog antes de persistir.
//
// Exemplo:
//
//	svc := backlog.NewService(repo, backlog.DefaultValidator())
type Validator interface {
	Validate(b Backlog) error
}

// StrictValidator implementa as regras de negócio padrão.
type StrictValidator struct{}

// DefaultValidator retorna um Validator com as regras de negócio padrão.
func DefaultValidator() Validator { return StrictValidator{} }

// Validate retorna *InputError se o Backlog violar alguma constraint.
func (StrictValidator) Validate(b Backlog) error {
	if strings.TrimSpace(b.Title) == "" {
		return &InputError{Kind: KindInvalid, Field: "title", Message: fmt.Sprintf("title must not be empty, got %q", b.Title)}
	}
	if len(b.Title) > MaxTitleLength {
		return &InputError{Kind: KindInvalid, Field: "title", Message: fmt.Sprintf("title must be at most %d characters, got %d", MaxTitleLength, len(b.Title))}
	}
	if strings.TrimSpace(b.Description) == "" {
		return &InputError{Kind: KindInvalid, Field: "description", Message: fmt.Sprintf("description must not be empty, got %q", b.Description)}
	}
	if len(b.Description) > MaxDescriptionLength {
		return &InputError{Kind: KindInvalid, Field: "description", Message: fmt.Sprintf("description must be at most %d characters, got %d", MaxDescriptionLength, len(b.Description))}
	}
	return nil
}
