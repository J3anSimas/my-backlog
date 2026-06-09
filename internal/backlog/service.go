package backlog

import "context"

// Service executa as regras de negócio de backlogs.
//
// Exemplo:
//
//	svc := backlog.NewService(repo, backlog.DefaultValidator())
//	b, err := svc.Create(ctx, "Sprint 1", "Backlog do primeiro sprint")
type Service struct {
	repo      Repository
	validator Validator
}

// NewService cria um Service com o repositório e validador injetados.
func NewService(repo Repository, v Validator) *Service {
	return &Service{repo: repo, validator: v}
}

// Create valida e persiste um novo Backlog com o título e descrição fornecidos.
func (s *Service) Create(ctx context.Context, title, description string) (Backlog, error) {
	if err := s.validator.Validate(Backlog{Title: title, Description: description}); err != nil {
		return Backlog{}, err
	}
	return s.repo.Save(ctx, Backlog{Title: title, Description: description})
}
