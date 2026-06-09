// Package database fornece conexão e utilitários para o banco Oracle.
package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	go_ora "github.com/sijms/go-ora/v2"
)

// Config reúne os parâmetros necessários para abrir uma conexão Oracle com wallet.
type Config struct {
	User        string
	Password    string
	Host        string
	Port        int
	ServiceName string
	// WalletPath é o diretório que contém os arquivos cwallet.sso / ewallet.p12.
	WalletPath string
}

// ConfigFromEnv lê a configuração das variáveis de ambiente:
//
//	DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_SERVICE_NAME, DB_WALLET_PATH
func ConfigFromEnv() Config {
	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	if port == 0 {
		port = 1521
	}
	return Config{
		User:        os.Getenv("DB_USER"),
		Password:    os.Getenv("DB_PASSWORD"),
		Host:        os.Getenv("DB_HOST"),
		Port:        port,
		ServiceName: os.Getenv("DB_SERVICE_NAME"),
		WalletPath:  os.Getenv("DB_WALLET_PATH"),
	}
}

// NewOracleDB abre e valida uma conexão Oracle usando TLS com wallet.
func NewOracleDB(cfg Config) (*sql.DB, error) {
	connStr := go_ora.BuildUrl(cfg.Host, cfg.Port, cfg.ServiceName, cfg.User, cfg.Password, map[string]string{
		"WALLET":     cfg.WalletPath,
		"SSL":        "TRUE",
		"SSL Verify": "FALSE",
	})

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening oracle connection (host=%s service=%s): %w", cfg.Host, cfg.ServiceName, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging oracle database (host=%s service=%s): %w", cfg.Host, cfg.ServiceName, err)
	}

	return db, nil
}
