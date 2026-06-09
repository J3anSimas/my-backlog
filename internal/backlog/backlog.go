// Package backlog gerencia o cadastro de product backlogs de um ambiente.
package backlog

// Backlog representa um product backlog com título e descrição.
type Backlog struct {
	ID          string
	Title       string
	Description string
}
