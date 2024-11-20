package session

import (
	"sync"

	"github.com/google/uuid"
)

// Locks the given manager behind a mutex, allowing for concurrent usage
type MutexManager struct {
	mu      sync.RWMutex
	manager Manager
}

func NewMutexManager(m Manager) *MutexManager {
	return &MutexManager{
		manager: m,
	}
}

var (
	_ Manager = &MutexManager{}
)

func (m *MutexManager) NewSession(username string, id uuid.UUID) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.manager.NewSession(username, id)
}

func (m *MutexManager) GetSessionData(key string) (*SessionData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.manager.GetSessionData(key)
}

func (m *MutexManager) SetSessionData(key string, data *SessionData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.manager.SetSessionData(key, data)
}

func (m *MutexManager) RefreshSession(key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.manager.RefreshSession(key)
}

func (m *MutexManager) EndSession(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.manager.EndSession(key)
}
