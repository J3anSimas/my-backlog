// Package backlogtest fornece fakes para testes que dependem de backlog.Repository.
// Código de produção nunca deve importar este pacote.
package backlogtest

import (
	"fmt"

	"my-backlog/internal/backlog"
)

// FakeRepository implementa backlog.Repository em memória.
type FakeRepository struct {
	saved     []backlog.Backlog
	nextID    int
	failOnce  error
}

// NewFakeRepository cria um FakeRepository com estado limpo e IDs a partir de 1.
func NewFakeRepository() *FakeRepository {
	return &FakeRepository{}
}

// FailWith agenda um erro one-shot: a próxima chamada a Save retorna err e limpa o agendamento.
func (f *FakeRepository) FailWith(err error) {
	f.failOnce = err
}

// Save implementa backlog.Repository. Retorna erro one-shot se FailWith foi chamado.
func (f *FakeRepository) Save(b backlog.Backlog) (backlog.Backlog, error) {
	if f.failOnce != nil {
		err := f.failOnce
		f.failOnce = nil
		return backlog.Backlog{}, err
	}
	f.nextID++
	b.ID = fmt.Sprintf("%d", f.nextID)
	f.saved = append(f.saved, b)
	return b, nil
}

// Saved retorna uma cópia dos backlogs persistidos até o momento.
func (f *FakeRepository) Saved() []backlog.Backlog {
	cp := make([]backlog.Backlog, len(f.saved))
	copy(cp, f.saved)
	return cp
}

// SaveCount retorna quantas vezes Save foi chamado com sucesso.
func (f *FakeRepository) SaveCount() int {
	return len(f.saved)
}
