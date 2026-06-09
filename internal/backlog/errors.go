package backlog

import "fmt"

// InputKind classifica o motivo de um InputError.
type InputKind int

const (
	KindInvalid  InputKind = iota // HTTP 400
	KindNotFound                  // HTTP 404
)

// InputError é retornado quando os dados do chamador são inválidos.
// Nunca é retentável; indica que o chamador deve corrigir a entrada.
type InputError struct {
	Kind    InputKind
	Field   string // nome do campo com problema, ex.: "title"
	Message string // contém o valor ofensor e o formato esperado
}

func (e *InputError) Error() string { return e.Message }

// InfraError é retornado quando a infraestrutura falha (DB, geração de ID etc.).
// Pode ser retentável; indica problema operacional, não do chamador.
type InfraError struct {
	Op    string // operação que falhou, ex.: "db-insert"
	Cause error  // erro original, preservado via Unwrap
}

func (e *InfraError) Error() string { return fmt.Sprintf("%s: %v", e.Op, e.Cause) }
func (e *InfraError) Unwrap() error { return e.Cause }
