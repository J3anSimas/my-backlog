package database

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// fakeMigrationStore é um MigrationStore controlável para testes unitários.
type fakeMigrationStore struct {
	ensureErr  error
	appliedErr error
	applied    map[int]bool
	applyErr   error
	applyCalls int
	applyOrder []int
}

func (f *fakeMigrationStore) EnsureTracking(_ context.Context) error {
	return f.ensureErr
}

func (f *fakeMigrationStore) AppliedVersions(_ context.Context) (map[int]bool, error) {
	if f.appliedErr != nil {
		return nil, f.appliedErr
	}
	if f.applied == nil {
		return map[int]bool{}, nil
	}
	return f.applied, nil
}

func (f *fakeMigrationStore) Apply(_ context.Context, m Migration) error {
	f.applyCalls++
	f.applyOrder = append(f.applyOrder, m.Version)
	return f.applyErr
}

// --- Ciclo 1: tracer bullet ---

func TestUp_AllMigrationsApplied_DoesNothing(t *testing.T) {
	migrations := []Migration{
		{Version: 1, Name: "first", SQL: "CREATE TABLE t1"},
	}
	store := &fakeMigrationStore{applied: map[int]bool{1: true}}

	if err := NewMigrator(store, migrations).Up(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.applyCalls != 0 {
		t.Errorf("expected no Apply calls, got %d", store.applyCalls)
	}
}

// --- Ciclo 2: migrations pendentes são aplicadas em ordem ---

func TestUp_AppliesPendingMigrationsInOrder(t *testing.T) {
	migrations := []Migration{
		{Version: 1, Name: "first"},
		{Version: 2, Name: "second"},
	}
	store := &fakeMigrationStore{applied: map[int]bool{}}

	if err := NewMigrator(store, migrations).Up(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.applyCalls != 2 {
		t.Errorf("expected 2 Apply calls, got %d", store.applyCalls)
	}
	if store.applyOrder[0] != 1 || store.applyOrder[1] != 2 {
		t.Errorf("expected order [1 2], got %v", store.applyOrder)
	}
}

// --- Ciclo 3: falha em EnsureTracking para antes de aplicar ---

func TestUp_EnsureTrackingFails_StopsBeforeApply(t *testing.T) {
	store := &fakeMigrationStore{ensureErr: errors.New("tracking table missing")}

	err := NewMigrator(store, []Migration{{Version: 1}}).Up(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if store.applyCalls != 0 {
		t.Errorf("expected no Apply calls, got %d", store.applyCalls)
	}
}

// --- Ciclo 4: split-brain é propagado e detectável ---

func TestUp_SplitBrainError_IsPropagated(t *testing.T) {
	store := &fakeMigrationStore{
		applyErr: fmt.Errorf("connection lost: %w", ErrDDLAppliedUntracked),
	}
	migrations := []Migration{{Version: 1, Name: "first"}}

	err := NewMigrator(store, migrations).Up(context.Background())
	if !errors.Is(err, ErrDDLAppliedUntracked) {
		t.Errorf("expected ErrDDLAppliedUntracked in error chain, got %v", err)
	}
}

func TestParseMigrationFilename_ValidName(t *testing.T) {
	version, name, err := parseMigrationFilename("0001_create_backlogs.sql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
	if name != "create_backlogs" {
		t.Errorf("expected name %q, got %q", "create_backlogs", name)
	}
}

func TestParseMigrationFilename_NameWithUnderscores(t *testing.T) {
	version, name, err := parseMigrationFilename("0042_add_user_id_to_backlogs.sql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != 42 {
		t.Errorf("expected version 42, got %d", version)
	}
	if name != "add_user_id_to_backlogs" {
		t.Errorf("expected name %q, got %q", "add_user_id_to_backlogs", name)
	}
}

func TestParseMigrationFilename_MissingUnderscore_ReturnsError(t *testing.T) {
	_, _, err := parseMigrationFilename("0001create_backlogs.sql")
	if err == nil {
		t.Fatal("expected error for filename without underscore separator, got nil")
	}
}

func TestParseMigrationFilename_NonNumericVersion_ReturnsError(t *testing.T) {
	_, _, err := parseMigrationFilename("abc_create_backlogs.sql")
	if err == nil {
		t.Fatal("expected error for non-numeric version, got nil")
	}
}

func TestNormalizeMigrationSQL_StripsTrailingSemicolon(t *testing.T) {
	got := normalizeMigrationSQL("CREATE TABLE foo (id NUMBER);")
	want := "CREATE TABLE foo (id NUMBER)"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestNormalizeMigrationSQL_StripsLeadingAndTrailingWhitespace(t *testing.T) {
	got := normalizeMigrationSQL("  \n  CREATE TABLE foo (id NUMBER)  \n  ")
	want := "CREATE TABLE foo (id NUMBER)"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestPendingMigrations_ExcludesAppliedVersions(t *testing.T) {
	m := &Migrator{
		migrations: []Migration{
			{Version: 1, Name: "first"},
			{Version: 2, Name: "second"},
			{Version: 3, Name: "third"},
		},
	}

	applied := map[int]bool{1: true, 3: true}
	pending := m.pendingMigrations(applied)

	if len(pending) != 1 {
		t.Fatalf("expected 1 pending migration, got %d", len(pending))
	}
	if pending[0].Version != 2 {
		t.Errorf("expected pending version 2, got %d", pending[0].Version)
	}
}

func TestPendingMigrations_AllApplied_ReturnsEmpty(t *testing.T) {
	m := &Migrator{
		migrations: []Migration{
			{Version: 1, Name: "first"},
			{Version: 2, Name: "second"},
		},
	}

	applied := map[int]bool{1: true, 2: true}
	pending := m.pendingMigrations(applied)

	if len(pending) != 0 {
		t.Errorf("expected 0 pending migrations, got %d", len(pending))
	}
}

func TestPendingMigrations_NoneApplied_ReturnsAll(t *testing.T) {
	m := &Migrator{
		migrations: []Migration{
			{Version: 1, Name: "first"},
			{Version: 2, Name: "second"},
		},
	}

	pending := m.pendingMigrations(map[int]bool{})

	if len(pending) != 2 {
		t.Errorf("expected 2 pending migrations, got %d", len(pending))
	}
}
