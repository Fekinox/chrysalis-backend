package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// In-memory session manager. All sessions will end once the server closes, and
// it only recognizes sessions that connect to the same machine, so this is not
// ideal for use in production.
type MemorySessionManager struct {
	mu               sync.RWMutex
	sessions         map[string]*SessionData
	sessionKeyLength int
}

func NewMemorySessionManager() *MemorySessionManager {
	return &MemorySessionManager{
		sessions:         make(map[string]*SessionData),
		sessionKeyLength: 128,
	}
}

func (sm *MemorySessionManager) NewSession(
	username string,
	id uuid.UUID,
) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key, err := GenerateSessionKey()
	if err != nil {
		return "", err
	}

	for {
		if _, ok := sm.sessions[key]; !ok {
			break
		}
		key, err = GenerateSessionKey()
		if err != nil {
			return "", err
		}
	}

	sm.sessions[string(key)] = &SessionData{
		Username:  username,
		UserID:    id,
		CreatedAt: time.Now(),
	}

	return key, nil
}

func (sm *MemorySessionManager) GetSessionData(
	key string,
) (*SessionData, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data, ok := sm.sessions[key]
	if !ok {
		return nil, SessionNotFound
	}
	return data, nil
}

func (sm *MemorySessionManager) SetSessionData(
	key string,
	data *SessionData,
) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, ok := sm.sessions[key]
	if !ok {
		return SessionNotFound
	}
	sm.sessions[key] = data

	return nil
}

func (sm *MemorySessionManager) RefreshSession(key string) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, ok := sm.sessions[key]
	if !ok {
		return "", SessionNotFound
	}

	newKey := uuid.NewString()

	for {
		if _, ok := sm.sessions[newKey]; !ok {
			break
		}
		newKey = uuid.NewString()
	}

	sm.sessions[newKey] = sm.sessions[key]
	delete(sm.sessions, key)

	return newKey, nil
}

func (sm *MemorySessionManager) EndSession(key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, ok := sm.sessions[key]
	if !ok {
		return SessionNotFound
	}

	delete(sm.sessions, key)

	return nil
}
