// Package database fornece conexão e utilitários para o banco Oracle.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

// ErrNotConfigured is returned by Connect when the required env vars (at minimum DB_USER)
// are not set and no WithOpener was provided.
var ErrNotConfigured = errors.New("database: required env vars not set")

const (
	defaultConnectTimeout = 30 * time.Second
	defaultMigrateTimeout = 60 * time.Second
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

// oracleOpener é o Opener padrão que usa go-ora para conexões Oracle com wallet TLS.
type oracleOpener struct{}

func (oracleOpener) Open(ctx context.Context, cfg Config) (*sql.DB, error) {
	connStr := go_ora.BuildUrl(cfg.Host, cfg.Port, cfg.ServiceName, cfg.User, cfg.Password, map[string]string{
		"WALLET":     cfg.WalletPath,
		"SSL":        "TRUE",
		"SSL Verify": "FALSE",
	})

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening oracle connection (host=%s service=%s): %w", cfg.Host, cfg.ServiceName, err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging oracle database (host=%s service=%s): %w", cfg.Host, cfg.ServiceName, err)
	}

	return db, nil
}

// Connect opens and initializes a database connection using functional options.
// All options have safe defaults. Returns ErrNotConfigured when DB_USER is not set
// and no WithOpener was provided, allowing tests to call t.Skip cleanly.
//
// Integration test:
//
//	db, err := database.Connect(
//	    database.WithConfig(cfg),
//	    database.WithMigrations(database.MigrationSourceFunc(database.LoadMigrations)),
//	)
//	if errors.Is(err, database.ErrNotConfigured) { t.Skip("Oracle env vars não configurados") }
//
// Unit test without Oracle:
//
//	db, err := database.Connect(database.WithOpener(myFakeOpener{}))
func Connect(opts ...ConnectOption) (*sql.DB, error) {
	o := &connectOptions{
		connectTimeout: defaultConnectTimeout,
		migrateTimeout: defaultMigrateTimeout,
	}
	for _, opt := range opts {
		opt(o)
	}

	cfg := ConfigFromEnv()
	if o.cfg != nil {
		cfg = *o.cfg
	}

	opener := o.opener
	if opener == nil {
		if cfg.User == "" {
			return nil, ErrNotConfigured
		}
		opener = oracleOpener{}
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), o.connectTimeout)
	defer connectCancel()

	db, err := opener.Open(connectCtx, cfg)
	if err != nil {
		return nil, err
	}

	if o.migrationSource == nil {
		return db, nil
	}

	migrations, err := o.migrationSource.Migrations()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("loading migrations: %w", err)
	}

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), o.migrateTimeout)
	defer migrateCancel()

	if err := NewMigrator(NewOracleMigrationStore(db), migrations).Up(migrateCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("applying migrations: %w", err)
	}

	return db, nil
}

// OpenForApp opens a production database connection using env var config and embedded migrations.
// It is the recommended entry point for main().
//
// Example:
//
//	db, err := database.OpenForApp()
//	if err != nil { log.Fatalf("database: %v", err) }
//	defer db.Close()
func OpenForApp() (*sql.DB, error) {
	return Connect(
		WithConfig(ConfigFromEnv()),
		WithMigrations(MigrationSourceFunc(LoadMigrations)),
	)
}
