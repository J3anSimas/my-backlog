package backlog_test

import (
	"fmt"
	"strings"
	"testing"

	"my-backlog/internal/backlog"
)

// fakeRepository é um fake em memória que atribui IDs sequenciais.
type fakeRepository struct {
	saved  []backlog.Backlog
	nextID int
}

func (f *fakeRepository) Save(b backlog.Backlog) (backlog.Backlog, error) {
	f.nextID++
	b.ID = fmt.Sprintf("%d", f.nextID)
	f.saved = append(f.saved, b)
	return b, nil
}

func newService() (*backlog.Service, *fakeRepository) {
	repo := &fakeRepository{}
	return backlog.NewService(repo), repo
}

// --- Happy path ---

func TestService_Create_ReturnsBacklogWithTitleDescriptionAndID(t *testing.T) {
	svc, _ := newService()

	got, err := svc.Create("Meu Backlog", "Descrição do backlog de produto")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.Title != "Meu Backlog" {
		t.Errorf("título esperado %q, obtido %q", "Meu Backlog", got.Title)
	}
	if got.Description != "Descrição do backlog de produto" {
		t.Errorf("descrição esperada %q, obtida %q", "Descrição do backlog de produto", got.Description)
	}
	if got.ID == "" {
		t.Error("esperado ID não vazio")
	}
}

func TestService_Create_PersistsBacklogInRepository(t *testing.T) {
	svc, repo := newService()

	_, err := svc.Create("Backlog A", "Descrição A")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Errorf("esperado 1 backlog salvo, obtido %d", len(repo.saved))
	}
}

// --- Edge cases: título ---

func TestService_Create_EmptyTitle_ReturnsError(t *testing.T) {
	svc, _ := newService()

	_, err := svc.Create("", "Descrição válida")
	if err == nil {
		t.Fatal("esperado erro para título vazio, obtido nil")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("mensagem de erro deve mencionar 'title', obtido: %q", err.Error())
	}
}

func TestService_Create_WhitespaceOnlyTitle_ReturnsError(t *testing.T) {
	svc, _ := newService()

	_, err := svc.Create("   ", "Descrição válida")
	if err == nil {
		t.Fatal("esperado erro para título apenas com espaços, obtido nil")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("mensagem de erro deve mencionar 'title', obtido: %q", err.Error())
	}
}

func TestService_Create_TitleExceedsMaxLength_ReturnsError(t *testing.T) {
	svc, _ := newService()

	longTitle := strings.Repeat("a", 201)
	_, err := svc.Create(longTitle, "Descrição válida")
	if err == nil {
		t.Fatal("esperado erro para título acima de 200 caracteres, obtido nil")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("mensagem de erro deve mencionar 'title', obtido: %q", err.Error())
	}
}

// --- Edge cases: descrição ---

func TestService_Create_EmptyDescription_ReturnsError(t *testing.T) {
	svc, _ := newService()

	_, err := svc.Create("Título válido", "")
	if err == nil {
		t.Fatal("esperado erro para descrição vazia, obtido nil")
	}
	if !strings.Contains(err.Error(), "description") {
		t.Errorf("mensagem de erro deve mencionar 'description', obtido: %q", err.Error())
	}
}

func TestService_Create_DescriptionExceedsMaxLength_ReturnsError(t *testing.T) {
	svc, _ := newService()

	longDesc := strings.Repeat("a", 1001)
	_, err := svc.Create("Título válido", longDesc)
	if err == nil {
		t.Fatal("esperado erro para descrição acima de 1000 caracteres, obtido nil")
	}
	if !strings.Contains(err.Error(), "description") {
		t.Errorf("mensagem de erro deve mencionar 'description', obtido: %q", err.Error())
	}
}
