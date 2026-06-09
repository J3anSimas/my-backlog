// Package user gerencia o cadastro e autenticação de usuários da plataforma.
package user

import "time"

// User representa uma conta de usuário registrada.
type User struct {
	ID           string
	Name         string
	Email        string
	Password     string // texto plano — presente apenas durante o cadastro, nunca persistido
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
