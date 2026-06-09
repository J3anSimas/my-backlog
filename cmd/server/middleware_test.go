package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeSessionGetter implementa SessionGetter para testes de middleware.
type fakeSessionGetter struct {
	valid map[string]string
}

func (f *fakeSessionGetter) Get(sid string) (string, bool) {
	userID, ok := f.valid[sid]
	return userID, ok
}

func newFakeGetter(sids ...string) *fakeSessionGetter {
	m := make(map[string]string, len(sids))
	for _, sid := range sids {
		m[sid] = "user-1"
	}
	return &fakeSessionGetter{valid: m}
}

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// --- RequireAuth ---

// --- RedirectIfAuth ---

func TestRedirectIfAuth_ValidSid_RedirectsToHome(t *testing.T) {
	store := newFakeGetter("sid-abc")
	h := RedirectIfAuth(store, okHandler)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: "sid-abc"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/home" {
		t.Errorf("esperado redirect para /home, obtido %q", rec.Header().Get("Location"))
	}
}

func TestRedirectIfAuth_NoCookie_CallsNext(t *testing.T) {
	store := newFakeGetter()
	h := RedirectIfAuth(store, okHandler)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obtido %d", rec.Code)
	}
}

func TestRequireAuth_ValidSid_CallsNext(t *testing.T) {
	store := newFakeGetter("sid-abc")
	h := RequireAuth(store, okHandler)

	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: "sid-abc"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obtido %d", rec.Code)
	}
}

func TestRequireAuth_InvalidSid_RedirectsToLogin(t *testing.T) {
	store := newFakeGetter() // nenhum sid válido
	h := RequireAuth(store, okHandler)

	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: "sid-invalido"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/login" {
		t.Errorf("esperado redirect para /login, obtido %q", rec.Header().Get("Location"))
	}
}

func TestRequireAuth_NoCookie_RedirectsToLogin(t *testing.T) {
	store := newFakeGetter()
	h := RequireAuth(store, okHandler)

	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/login" {
		t.Errorf("esperado redirect para /login, obtido %q", rec.Header().Get("Location"))
	}
}
