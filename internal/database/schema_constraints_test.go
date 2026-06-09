package database

import (
	"fmt"
	"io/fs"
	"regexp"
	"strconv"
	"testing"

	"my-backlog/internal/backlog"
)

func TestSchemaConstraintsMatchGoConstants(t *testing.T) {
	sql := mustReadMigration(t, "0001_create_backlogs.sql")
	assertVarchar2Limit(t, sql, "title", backlog.MaxTitleLength)
	assertVarchar2Limit(t, sql, "description", backlog.MaxDescriptionLength)
}

func mustReadMigration(t *testing.T, filename string) string {
	t.Helper()
	content, err := fs.ReadFile(migrationsFS, "migrations/"+filename)
	if err != nil {
		t.Fatalf("lendo migration %q: %v", filename, err)
	}
	return string(content)
}

// assertVarchar2Limit verifica que `column VARCHAR2(N)` no SQL bate com goLimit.
// Usa regex para tolerar espaços múltiplos entre nome da coluna e o tipo.
func assertVarchar2Limit(t *testing.T, sql, column string, goLimit int) {
	t.Helper()
	pattern := fmt.Sprintf(`(?i)\b%s\s+VARCHAR2\((\d+)\)`, regexp.QuoteMeta(column))
	re := regexp.MustCompile(pattern)

	m := re.FindStringSubmatch(sql)
	if m == nil {
		t.Errorf("coluna %q não encontrada no SQL como VARCHAR2", column)
		return
	}

	sqlLimit, err := strconv.Atoi(m[1])
	if err != nil {
		t.Errorf("limite de VARCHAR2 não é inteiro para coluna %q: %q", column, m[1])
		return
	}

	if sqlLimit != goLimit {
		t.Errorf("drift detectado em %q: SQL=VARCHAR2(%d), Go=%d — atualize os dois juntos",
			column, sqlLimit, goLimit)
	}
}
