package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"my-backlog/internal/session"
	"my-backlog/internal/user"
)

//go:embed templates
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/*.html"))

// buildMux monta o ServeMux com todas as rotas da aplicação.
func buildMux(svc *user.Service, store session.Sessions) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /{$}", RedirectIfAuth(store, http.HandlerFunc(getRoot)))
	mux.HandleFunc("GET /register", getRegister)
	mux.HandleFunc("POST /register", postRegister(svc, store))
	mux.Handle("GET /login", RedirectIfAuth(store, http.HandlerFunc(getLogin)))
	mux.HandleFunc("POST /login", postLogin(svc, store))
	mux.Handle("GET /home", RequireAuth(store, http.HandlerFunc(getHome)))
	return mux
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func getRegister(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "register.html", nil); err != nil {
		log.Printf("register template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func postRegister(svc *user.Service, store session.Sessions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		u, err := svc.Register(r.Context(),
			r.FormValue("name"),
			r.FormValue("email"),
			r.FormValue("password"),
		)
		if err != nil {
			writeError(w, err)
			return
		}

		sid := store.New(u.ID)
		http.SetCookie(w, &http.Cookie{
			Name:     "sid",
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
		log.Printf("login template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func postLogin(svc *user.Service, store session.Sessions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		u, err := svc.Login(r.Context(), r.FormValue("email"), r.FormValue("password"))
		if err != nil {
			writeError(w, err)
			return
		}

		sid := store.New(u.ID)
		http.SetCookie(w, &http.Cookie{
			Name:     "sid",
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func getHome(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		log.Printf("home template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

