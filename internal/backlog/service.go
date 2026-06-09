package backlog

import (
	"context"
	"fmt"
	"strings"
)

// Service executa as regras de negócio de backlogs.
//
// Exemplo:
//
//	svc := backlog.NewService(repo)
//	b, err := svc.Create(ctx, "Sprint 1", "Backlog do primeiro sprint")
type Service struct {
	repo Repository
}

// NewService cria um Service com o repositório injetado.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create valida e persiste um novo Backlog com o título e descrição fornecidos.
func (s *Service) Create(ctx context.Context, title, description string) (Backlog, error) {
	if err := validateTitle(title); err != nil {
		return Backlog{}, err
	}
	if err := validateDescription(description); err != nil {
		return Backlog{}, err
	}
	return s.repo.Save(ctx, Backlog{Title: title, Description: description})
}

func validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title must not be empty, got %q", title)
	}
	if len(title) > 200 {
		return fmt.Errorf("title must be at most %d characters, got %d", 200, len(title))
	}
	return nil
}

func validateDescription(description string) error {
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description must not be empty, got %q", description)
	}
	if len(description) > 1000 {
		return fmt.Errorf("description must be at most %d characters, got %d", 1000, len(description))
	}
	return nil
}
