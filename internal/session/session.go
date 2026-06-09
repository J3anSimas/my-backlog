// Package session gerencia sessões de usuário em memória.
// Sessões são perdidas ao reiniciar o processo — aceitável para esta fase.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// Store mantém o mapa de session ID → userID em memória.
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewStore cria um Store vazio.
func NewStore() *Store {
	return &Store{data: make(map[string]string)}
}

// New cria uma nova sessão para o userID e retorna o session ID opaco.
func (s *Store) New(userID string) string {
	sid := newSessionID()
	s.mu.Lock()
	s.data[sid] = userID
	s.mu.Unlock()
	return sid
}

// Get retorna o userID associado ao sid, e false se a sessão não existir.
func (s *Store) Get(sid string) (string, bool) {
	s.mu.RLock()
	userID, ok := s.data[sid]
	s.mu.RUnlock()
	return userID, ok
}

// Delete remove a sessão identificada por sid.
func (s *Store) Delete(sid string) {
	s.mu.Lock()
	delete(s.data, sid)
	s.mu.Unlock()
}

func newSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("session: rand.Read failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}
