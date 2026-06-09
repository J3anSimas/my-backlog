package backlog

import (
	"fmt"
	"strings"
)

const (
	maxTitleLength       = 200
	maxDescriptionLength = 1000
)

// Service executa as regras de negócio de backlogs.
//
// Exemplo:
//
//	svc := backlog.NewService(repo)
//	b, err := svc.Create("Sprint 1", "Backlog do primeiro sprint")
type Service struct {
	repo Repository
}

// NewService cria um Service com o repositório injetado.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create valida e persiste um novo Backlog com o título e descrição fornecidos.
func (s *Service) Create(title, description string) (Backlog, error) {
	if err := validateTitle(title); err != nil {
		return Backlog{}, err
	}
	if err := validateDescription(description); err != nil {
		return Backlog{}, err
	}
	return s.repo.Save(Backlog{Title: title, Description: description})
}

func validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title must not be empty, got %q", title)
	}
	if len(title) > maxTitleLength {
		return fmt.Errorf("title must be at most %d characters, got %d", maxTitleLength, len(title))
	}
	return nil
}

func validateDescription(description string) error {
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description must not be empty, got %q", description)
	}
	if len(description) > maxDescriptionLength {
		return fmt.Errorf("description must be at most %d characters, got %d", maxDescriptionLength, len(description))
	}
	return nil
}
