package backlog

// Repository persiste e recupera Backlogs.
//
// Exemplo:
//
//	repo := NewPostgresRepository(db)
//	saved, err := repo.Save(b)
type Repository interface {
	Save(b Backlog) (Backlog, error)
}
