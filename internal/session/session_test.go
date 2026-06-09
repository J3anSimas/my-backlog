package session_test

import (
	"testing"

	"my-backlog/internal/session"
)

func TestStore_NewAndGet_ReturnsSameUserID(t *testing.T) {
	store := session.NewStore()

	sid := store.New("user-42")

	if sid == "" {
		t.Fatal("esperado session ID não vazio")
	}
	got, ok := store.Get(sid)
	if !ok {
		t.Fatal("esperado ok=true para session ID válido")
	}
	if got != "user-42" {
		t.Errorf("esperado userID=%q, obtido %q", "user-42", got)
	}
}

func TestStore_Get_UnknownID_ReturnsFalse(t *testing.T) {
	store := session.NewStore()

	_, ok := store.Get("id-inexistente")

	if ok {
		t.Error("esperado ok=false para session ID desconhecido")
	}
}

func TestStore_Delete_RemovesSession(t *testing.T) {
	store := session.NewStore()
	sid := store.New("user-99")

	store.Delete(sid)

	_, ok := store.Get(sid)
	if ok {
		t.Error("esperado ok=false após Delete")
	}
}

func TestStore_New_EachCallReturnsDistinctID(t *testing.T) {
	store := session.NewStore()

	sid1 := store.New("user-1")
	sid2 := store.New("user-2")

	if sid1 == sid2 {
		t.Error("esperado IDs de sessão distintos")
	}
}
