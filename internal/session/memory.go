package session

import (
	"sync"
)

// In-memory session manager. All sessions will end once the server closes, and
// it only recognizes sessions that connect to the same machine, so this is not
// ideal for use in production.
type MemoryBackend struct {
	mu       sync.RWMutex
	sessions map[string]*SessionData
}

func NewMemorySessionManager() *Manager {
	return &Manager{&MemoryBackend{
		sessions: make(map[string]*SessionData),
	}}
}

func (m *MemoryBackend) Set(key string, value *SessionData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[key] = value
	return nil
}

func (m *MemoryBackend) Get(key string) (*SessionData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.sessions[key]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return v, nil
}

func (m *MemoryBackend) Del(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, key)

	return nil
}
