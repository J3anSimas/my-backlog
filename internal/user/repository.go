package user

import "context"

// Repository persiste e consulta Users.
//
// Exemplo:
//
//	repo := user.NewOracleRepository(db)
//	saved, err := repo.Save(ctx, u)
type Repository interface {
	Save(ctx context.Context, u User) (User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}
