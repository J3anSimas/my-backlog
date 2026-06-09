package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"my-backlog/internal/database"
	"my-backlog/internal/session"
	"my-backlog/internal/user"
)

func main() {
	// .env é opcional: ausente em produção (vars já no ambiente), presente no dev local.
	// Falha apenas se o arquivo existir mas tiver erro de sintaxe.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf(".env: %v", err)
	}

	db, err := database.OpenForApp()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	log.Println("database ready")

	userRepo := user.NewOracleRepository(db)
	userSvc := user.NewService(userRepo, user.DefaultValidator())
	sessionStore := session.NewStore()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8081"
	}

	mux := buildMux(userSvc, sessionStore)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
