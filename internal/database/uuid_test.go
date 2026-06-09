package database_test

import (
	"regexp"
	"testing"

	"my-backlog/internal/database"
)

// uuidV4 valida formato e bits obrigatórios do UUID v4 / variante RFC 4122.
var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewUUID_EachCallReturnsDistinctValue(t *testing.T) {
	id1, _ := database.NewUUID()
	id2, _ := database.NewUUID()

	if id1 == id2 {
		t.Errorf("esperado UUIDs distintos, ambos foram %q", id1)
	}
}

func TestNewUUID_ReturnsValidV4Format(t *testing.T) {
	id, err := database.NewUUID()

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !uuidV4.MatchString(id) {
		t.Errorf("UUID fora do formato v4: %q", id)
	}
}
