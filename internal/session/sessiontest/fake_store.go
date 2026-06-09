// Package sessiontest fornece fakes da interface session.Sessions para uso em testes.
package sessiontest

import "fmt"

// FakeStore é uma implementação determinística de session.Sessions.
// New retorna IDs sequenciais ("sid-1", "sid-2", …) em vez de usar crypto/rand.
type FakeStore struct {
	counter int
	data    map[string]string
}

// NewFakeStore cria um FakeStore vazio.
func NewFakeStore() *FakeStore {
	return &FakeStore{data: make(map[string]string)}
}

// New cria uma sessão e retorna um ID determinístico.
func (f *FakeStore) New(userID string) string {
	f.counter++
	sid := fmt.Sprintf("sid-%d", f.counter)
	f.data[sid] = userID
	return sid
}

// Get retorna o userID associado ao sid, e false se não existir.
func (f *FakeStore) Get(sid string) (string, bool) {
	userID, ok := f.data[sid]
	return userID, ok
}

// Delete remove a sessão identificada por sid.
func (f *FakeStore) Delete(sid string) {
	delete(f.data, sid)
}

// SeedSID pré-semeia uma sessão com sid e userID conhecidos,
// útil para preparar estado antes de fazer uma requisição.
func (f *FakeStore) SeedSID(sid, userID string) {
	f.data[sid] = userID
}
