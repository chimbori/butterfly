package core

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	ID        string
	CreatedAt time.Time
}

type SessionStore interface {
	Create() (string, error)
	IsValid(id string) bool
	Delete(id string)
	CleanupExpired() int
}

type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
	ttl      time.Duration
}

func NewInMemorySessionStore(ttl time.Duration) SessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]Session),
		ttl:      ttl,
	}
}

// Create generates a new session ID, stores it, and returns the ID.
func (s *InMemorySessionStore) Create() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[id] = Session{
		ID:        id,
		CreatedAt: time.Now(),
	}
	return id, nil
}

// IsValid checks if a session ID exists and is not expired.
func (s *InMemorySessionStore) IsValid(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		return false
	}

	if time.Since(session.CreatedAt) > s.ttl {
		return false
	}

	return true
}

func (s *InMemorySessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

// CleanupExpired removes all expired sessions from the map.
func (s *InMemorySessionStore) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	now := time.Now()
	for id, session := range s.sessions {
		if now.Sub(session.CreatedAt) > s.ttl {
			delete(s.sessions, id)
			count++
		}
	}
	return count
}
