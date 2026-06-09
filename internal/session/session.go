// Package session gerencia sessões de usuário em memória.
// Sessões são perdidas ao reiniciar o processo — aceitável para esta fase.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// Sessions é a interface que os handlers usam para criar, consultar e remover sessões.
// Qualquer implementação (MemoryStore, Redis, JWT) satisfaz esta interface.
type Sessions interface {
	New(userID string) string
	Get(sid string) (userID string, ok bool)
	Delete(sid string)
}

// MemoryStore mantém o mapa de session ID → userID em memória.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewMemoryStore cria um MemoryStore vazio.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]string)}
}

// New cria uma nova sessão para o userID e retorna o session ID opaco.
func (s *MemoryStore) New(userID string) string {
	sid := newSessionID()
	s.mu.Lock()
	s.data[sid] = userID
	s.mu.Unlock()
	return sid
}

// Get retorna o userID associado ao sid, e false se a sessão não existir.
func (s *MemoryStore) Get(sid string) (string, bool) {
	s.mu.RLock()
	userID, ok := s.data[sid]
	s.mu.RUnlock()
	return userID, ok
}

// Delete remove a sessão identificada por sid.
func (s *MemoryStore) Delete(sid string) {
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
