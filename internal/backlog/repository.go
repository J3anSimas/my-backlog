package backlog

import "context"

// Repository persiste e recupera Backlogs.
//
// Exemplo:
//
//	repo := NewPostgresRepository(db)
//	saved, err := repo.Save(ctx, b)
type Repository interface {
	Save(ctx context.Context, b Backlog) (Backlog, error)
}
