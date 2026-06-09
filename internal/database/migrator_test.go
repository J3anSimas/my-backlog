package database

import (
	"testing"
)

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
