package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"my-backlog/internal/session/sessiontest"
	"my-backlog/internal/user"
	"my-backlog/internal/user/usertest"
)

func newTestServer() (*http.ServeMux, *usertest.FakeRepository, *sessiontest.FakeStore) {
	repo := usertest.NewFakeRepository()
	svc := user.NewService(repo, user.DefaultValidator())
	store := sessiontest.NewFakeStore()
	mux := buildMux(svc, store)
	return mux, repo, store
}

// --- GET / ---

func TestGetRoot_WithoutSession_RedirectsToLogin(t *testing.T) {
	mux, _, _ := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/login" {
		t.Errorf("esperado redirect para /login, obtido %q", rec.Header().Get("Location"))
	}
}

func TestGetRoot_WithValidSession_RedirectsToHome(t *testing.T) {
	mux, _, store := newTestServer()
	sid := store.New("user-1")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/home" {
		t.Errorf("esperado redirect para /home, obtido %q", rec.Header().Get("Location"))
	}
}

// --- GET /register ---

func TestGetRegister_Returns200WithForm(t *testing.T) {
	mux, _, _ := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obtido %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "register") {
		t.Error("esperado body contendo 'register'")
	}
}

// --- POST /register: caminho feliz ---

func TestPostRegister_ValidData_RedirectsToHome(t *testing.T) {
	mux, _, _ := newTestServer()

	form := url.Values{
		"name":     {"João Silva"},
		"email":    {"joao@email.com"},
		"password": {"senha123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/home" {
		t.Errorf("esperado redirect para /home, obtido %q", rec.Header().Get("Location"))
	}
}

func TestPostRegister_ValidData_SetsCookie(t *testing.T) {
	mux, _, _ := newTestServer()

	form := url.Values{
		"name":     {"João Silva"},
		"email":    {"joao@email.com"},
		"password": {"senha123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "sid" {
			found = true
			break
		}
	}
	if !found {
		t.Error("esperado cookie 'sid' na resposta")
	}
}

// --- POST /register: dados inválidos ---

func TestPostRegister_EmptyName_Returns422WithFieldError(t *testing.T) {
	mux, _, _ := newTestServer()

	form := url.Values{
		"name":     {""},
		"email":    {"joao@email.com"},
		"password": {"senha123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("esperado 422, obtido %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("body não é JSON válido: %v", err)
	}
	if body["field"] != "name" {
		t.Errorf("esperado field=%q, obtido %q", "name", body["field"])
	}
}

func TestPostRegister_DuplicateEmail_Returns409(t *testing.T) {
	mux, _, _ := newTestServer()

	form := url.Values{
		"name":     {"João"},
		"email":    {"joao@email.com"},
		"password": {"senha123"},
	}
	post := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	post() // primeiro cadastro
	rec := post() // duplicado

	if rec.Code != http.StatusConflict {
		t.Errorf("esperado 409, obtido %d", rec.Code)
	}
}

// --- GET /login ---

func TestGetLogin_WithoutSession_Returns200(t *testing.T) {
	mux, _, _ := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obtido %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "login") {
		t.Error("esperado body contendo 'login'")
	}
}

func TestGetLogin_WithValidSession_RedirectsToHome(t *testing.T) {
	mux, _, store := newTestServer()
	sid := store.New("user-1")

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/home" {
		t.Errorf("esperado redirect para /home, obtido %q", rec.Header().Get("Location"))
	}
}

// --- POST /login ---

func registerForLogin(t *testing.T, mux *http.ServeMux, email, password string) {
	t.Helper()
	form := url.Values{"name": {"Usuário"}, "email": {email}, "password": {password}}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(httptest.NewRecorder(), req)
}

func TestPostLogin_ValidCredentials_RedirectsToHome(t *testing.T) {
	mux, _, _ := newTestServer()
	registerForLogin(t, mux, "joao@email.com", "senha123")

	form := url.Values{"email": {"joao@email.com"}, "password": {"senha123"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/home" {
		t.Errorf("esperado redirect para /home, obtido %q", rec.Header().Get("Location"))
	}
}

func TestPostLogin_ValidCredentials_SetsCookie(t *testing.T) {
	mux, _, _ := newTestServer()
	registerForLogin(t, mux, "joao@email.com", "senha123")

	form := url.Values{"email": {"joao@email.com"}, "password": {"senha123"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var found bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == "sid" {
			found = true
			break
		}
	}
	if !found {
		t.Error("esperado cookie 'sid' na resposta")
	}
}

func TestPostLogin_InvalidCredentials_Returns401(t *testing.T) {
	mux, _, _ := newTestServer()
	registerForLogin(t, mux, "joao@email.com", "senha123")

	form := url.Values{"email": {"joao@email.com"}, "password": {"senhaerrada"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, obtido %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("body não é JSON válido: %v", err)
	}
	if body["message"] == "" {
		t.Error("esperado message não vazia no body")
	}
	if body["field"] != "" {
		t.Errorf("esperado field vazio (genérico), obtido %q", body["field"])
	}
}

func TestPostLogin_UnknownEmail_Returns401(t *testing.T) {
	mux, _, _ := newTestServer()

	form := url.Values{"email": {"naoexiste@email.com"}, "password": {"senha123"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, obtido %d", rec.Code)
	}
}

// --- GET /home ---

func TestGetHome_WithoutSession_RedirectsToLogin(t *testing.T) {
	mux, _, _ := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("esperado 303, obtido %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/login" {
		t.Errorf("esperado redirect para /login, obtido %q", rec.Header().Get("Location"))
	}
}

func TestGetHome_WithValidSession_Returns200(t *testing.T) {
	mux, _, store := newTestServer()
	sid := store.New("user-1")

	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obtido %d", rec.Code)
	}
}
