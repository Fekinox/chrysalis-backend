package session

import (
	"sync"

	"github.com/google/uuid"
)

type MemorySessionManager struct {
	mu sync.RWMutex
	sessions map[SessionKey]*SessionData
	sessionKeyLength int
}

func NewMemorySessionManager() *MemorySessionManager {
	return &MemorySessionManager{
		sessions: make(map[SessionKey]*SessionData),
		sessionKeyLength: 128,
	}
}

func (sm *MemorySessionManager) NewSession(username string, id uuid.UUID) SessionKey {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := uuid.NewString()

	for {
		if _, ok := sm.sessions[SessionKey(key)]; !ok {
			break
		}
		key = uuid.NewString()
	}

	sm.sessions[SessionKey(key)] = &SessionData{
		Username: username,
		UserID: id,
	}

	return SessionKey(key)
}

func (sm *MemorySessionManager) GetSessionData(key SessionKey) (*SessionData, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data, ok := sm.sessions[key]
	if !ok {
		return nil, SessionNotFound
	}
	return data, nil
}

func (sm *MemorySessionManager) SetSessionData(key SessionKey, data *SessionData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, ok := sm.sessions[key]
	if !ok {
		return SessionNotFound
	}
	sm.sessions[key] = data

	return nil
}

func (sm *MemorySessionManager) RefreshSession(key SessionKey) (SessionKey, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, ok := sm.sessions[key]
	if !ok {
		return SessionKey(""), SessionNotFound
	}

	newKey := uuid.NewString()

	for {
		if _, ok := sm.sessions[SessionKey(newKey)]; !ok {
			break
		}
		newKey = uuid.NewString()
	}

	sm.sessions[SessionKey(newKey)] = sm.sessions[key]
	delete(sm.sessions, key)

	return SessionKey(key), nil
}

func (sm *MemorySessionManager) EndSession(key SessionKey) error {
	_, ok := sm.sessions[key]
	if !ok {
		return SessionNotFound
	}

	delete(sm.sessions, key)

	return nil
}
