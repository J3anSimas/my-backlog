package backlogtest_test

import (
	"errors"
	"testing"

	"my-backlog/internal/backlog"
	"my-backlog/internal/backlog/backlogtest"
)

func TestFakeRepository_Save_AssignsSequentialIDs(t *testing.T) {
	repo := backlogtest.NewFakeRepository()

	first, err := repo.Save(backlog.Backlog{Title: "A", Description: "desc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	second, err := repo.Save(backlog.Backlog{Title: "B", Description: "desc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if first.ID == "" {
		t.Error("primeiro ID não pode ser vazio")
	}
	if second.ID == "" {
		t.Error("segundo ID não pode ser vazio")
	}
	if first.ID == second.ID {
		t.Errorf("IDs devem ser únicos, ambos foram %q", first.ID)
	}
}

func TestFakeRepository_Saved_ReturnsCopyOfPersistedBacklogs(t *testing.T) {
	repo := backlogtest.NewFakeRepository()

	repo.Save(backlog.Backlog{Title: "A", Description: "desc A"}) //nolint:errcheck
	repo.Save(backlog.Backlog{Title: "B", Description: "desc B"}) //nolint:errcheck

	saved := repo.Saved()
	if len(saved) != 2 {
		t.Fatalf("esperado 2 itens, obtido %d", len(saved))
	}
	if saved[0].Title != "A" {
		t.Errorf("primeiro título esperado %q, obtido %q", "A", saved[0].Title)
	}
	if saved[1].Title != "B" {
		t.Errorf("segundo título esperado %q, obtido %q", "B", saved[1].Title)
	}

	// mutação na cópia não afeta o estado interno
	saved[0].Title = "MUTADO"
	if repo.Saved()[0].Title != "A" {
		t.Error("Saved() deve retornar cópia — mutação externa não pode afetar o estado interno")
	}
}

func TestFakeRepository_SaveCount_ReflectsSuccessfulSaves(t *testing.T) {
	repo := backlogtest.NewFakeRepository()

	if repo.SaveCount() != 0 {
		t.Fatalf("esperado 0 antes de qualquer Save, obtido %d", repo.SaveCount())
	}

	repo.Save(backlog.Backlog{Title: "A", Description: "desc"}) //nolint:errcheck
	repo.Save(backlog.Backlog{Title: "B", Description: "desc"}) //nolint:errcheck

	if repo.SaveCount() != 2 {
		t.Errorf("esperado 2 após dois saves, obtido %d", repo.SaveCount())
	}
}

func TestFakeRepository_FailWith_ReturnsErrorOnNextSaveOnly(t *testing.T) {
	repo := backlogtest.NewFakeRepository()
	dbErr := errors.New("db down")

	repo.FailWith(dbErr)

	_, err := repo.Save(backlog.Backlog{Title: "X", Description: "desc"})
	if !errors.Is(err, dbErr) {
		t.Fatalf("esperado %v, obtido %v", dbErr, err)
	}
	if repo.SaveCount() != 0 {
		t.Errorf("Save com erro não deve incrementar SaveCount, obtido %d", repo.SaveCount())
	}

	// one-shot: a chamada seguinte deve ter sucesso
	_, err = repo.Save(backlog.Backlog{Title: "Y", Description: "desc"})
	if err != nil {
		t.Errorf("segundo Save deve ter sucesso após one-shot, obtido: %v", err)
	}
	if repo.SaveCount() != 1 {
		t.Errorf("esperado 1 após save bem-sucedido, obtido %d", repo.SaveCount())
	}
}
