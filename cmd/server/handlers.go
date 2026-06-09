package main

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"

	"my-backlog/internal/apperrors"
	"my-backlog/internal/session"
	"my-backlog/internal/user"
)

//go:embed templates
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/*.html"))

// buildMux monta o ServeMux com todas as rotas da aplicação.
func buildMux(svc *user.Service, store *session.Store) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", getRoot(store))
	mux.HandleFunc("GET /register", getRegister)
	mux.HandleFunc("POST /register", postRegister(svc, store))
	mux.HandleFunc("GET /home", getHome(store))
	return mux
}

func getRoot(store *session.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sid")
		if err == nil {
			if _, ok := store.Get(cookie.Value); ok {
				http.Redirect(w, r, "/home", http.StatusSeeOther)
				return
			}
		}
		http.Redirect(w, r, "/register", http.StatusSeeOther)
	}
}

func getRegister(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "register.html", nil); err != nil {
		log.Printf("register template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func postRegister(svc *user.Service, store *session.Store) http.HandlerFunc {
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
			writeRegisterError(w, err)
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

func getHome(store *session.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sid")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		if _, ok := store.Get(cookie.Value); !ok {
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "home.html", nil); err != nil {
			log.Printf("home template: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

type errorResponse struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func writeRegisterError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var inputErr *apperrors.InputError
	if errors.As(err, &inputErr) {
		status := http.StatusUnprocessableEntity
		if inputErr.Kind == apperrors.KindConflict {
			status = http.StatusConflict
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(errorResponse{Field: inputErr.Field, Message: inputErr.Message})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(errorResponse{Message: "internal error"})
}
