package main

import "net/http"

// SessionGetter é a fatia de session.Sessions que os middlewares precisam.
// Definida aqui (consumidora), não em internal/session, para manter o fluxo de dependência correto.
type SessionGetter interface {
	Get(sid string) (userID string, ok bool)
}

// RequireAuth redireciona para /login se a requisição não carrega uma sessão válida.
func RequireAuth(store SessionGetter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasSession(store, r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RedirectIfAuth redireciona para /home se a requisição já carrega uma sessão válida.
func RedirectIfAuth(store SessionGetter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasSession(store, r) {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func hasSession(store SessionGetter, r *http.Request) bool {
	cookie, err := r.Cookie("sid")
	if err != nil || cookie.Value == "" {
		return false
	}
	_, ok := store.Get(cookie.Value)
	return ok
}
