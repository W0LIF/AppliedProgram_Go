package storage

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]int
}

func NewTokenStore() *TokenStore {
	return &TokenStore{tokens: make(map[string]int)}
}

func (s *TokenStore) Issue(userID int) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("token: " + err.Error())
	}
	token := hex.EncodeToString(b)
	s.mu.Lock()
	s.tokens[token] = userID
	s.mu.Unlock()
	return token
}

func (s *TokenStore) Get(token string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.tokens[token]
	return id, ok
}

func (s *TokenStore) Revoke(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}
