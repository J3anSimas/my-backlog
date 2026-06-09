// Package usertest fornece fakes para testes que dependem de user.Repository.
// Código de produção nunca deve importar este pacote.
package usertest

import (
	"context"
	"fmt"

	"my-backlog/internal/user"
)

// FakeRepository implementa user.Repository em memória.
type FakeRepository struct {
	saved    []user.User
	nextID   int
	failOnce error
}

// NewFakeRepository cria um FakeRepository com estado limpo.
func NewFakeRepository() *FakeRepository { return &FakeRepository{} }

// FailWith agenda um erro one-shot: a próxima chamada a Save retorna err e limpa o agendamento.
func (f *FakeRepository) FailWith(err error) { f.failOnce = err }

// Save implementa user.Repository.
func (f *FakeRepository) Save(ctx context.Context, u user.User) (user.User, error) {
	if ctx.Err() != nil {
		return user.User{}, ctx.Err()
	}
	if f.failOnce != nil {
		err := f.failOnce
		f.failOnce = nil
		return user.User{}, err
	}
	f.nextID++
	u.ID = fmt.Sprintf("%d", f.nextID)
	f.saved = append(f.saved, u)
	return u, nil
}

// EmailExists retorna true se algum usuário salvo possui o e-mail informado.
func (f *FakeRepository) EmailExists(_ context.Context, email string) (bool, error) {
	for _, u := range f.saved {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

// Saved retorna uma cópia dos usuários persistidos até o momento.
func (f *FakeRepository) Saved() []user.User {
	cp := make([]user.User, len(f.saved))
	copy(cp, f.saved)
	return cp
}

// SaveCount retorna quantas vezes Save foi chamado com sucesso.
func (f *FakeRepository) SaveCount() int { return len(f.saved) }
