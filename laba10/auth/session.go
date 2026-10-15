package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Session struct {
	UserID int
	Role   string
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]Session)}
}

func (s *SessionStore) Create(userID int, role string) string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	s.mu.Lock()
	s.sessions[id] = Session{UserID: userID, Role: role}
	s.mu.Unlock()
	return id
}

func (s *SessionStore) Get(id string) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	return sess, ok
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}
