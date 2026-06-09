package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"my-backlog/internal/database"
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
}
