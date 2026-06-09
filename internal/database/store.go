package database

import (
	"context"
	"database/sql"
	"time"
)

// Opener opens and validates a database connection.
// Implement this interface to inject a fake connection in tests without a real Oracle instance.
//
// Example:
//
//	type myFakeOpener struct{}
//	func (myFakeOpener) Open(_ context.Context, _ database.Config) (*sql.DB, error) {
//	    return sql.Open("sqlite3", ":memory:")
//	}
type Opener interface {
	Open(ctx context.Context, cfg Config) (*sql.DB, error)
}

// MigrationSource provides the list of migrations to apply on Connect.
type MigrationSource interface {
	Migrations() ([]Migration, error)
}

// MigrationSourceFunc adapts a zero-argument function to MigrationSource.
//
// Example:
//
//	database.Connect(database.WithMigrations(database.MigrationSourceFunc(database.LoadMigrations)))
type MigrationSourceFunc func() ([]Migration, error)

func (f MigrationSourceFunc) Migrations() ([]Migration, error) { return f() }

// ConnectOption configures a Connect call.
type ConnectOption func(*connectOptions)

type connectOptions struct {
	cfg             *Config
	opener          Opener
	migrationSource MigrationSource
	connectTimeout  time.Duration
	migrateTimeout  time.Duration
}

// WithConfig overrides the configuration used to open the database.
// Without this option, Connect reads from environment variables.
func WithConfig(cfg Config) ConnectOption {
	return func(o *connectOptions) { o.cfg = &cfg }
}

// WithOpener replaces the default Oracle opener with a custom one.
// Use this in tests to inject a fake connection without a real Oracle instance.
func WithOpener(opener Opener) ConnectOption {
	return func(o *connectOptions) { o.opener = opener }
}

// WithMigrations sets the migration source to run automatically on Connect.
func WithMigrations(src MigrationSource) ConnectOption {
	return func(o *connectOptions) { o.migrationSource = src }
}

// WithConnectTimeout sets the maximum time to wait for the initial connection (default 30s).
func WithConnectTimeout(d time.Duration) ConnectOption {
	return func(o *connectOptions) { o.connectTimeout = d }
}

// WithMigrateTimeout sets the maximum time to wait for migrations to complete (default 60s).
func WithMigrateTimeout(d time.Duration) ConnectOption {
	return func(o *connectOptions) { o.migrateTimeout = d }
}
