package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/genbytes"
	"github.com/google/uuid"
)

const SESSION_KEY_LENGTH_BYTES = 32
const CSRF_TOKEN_LENGTH_BYTES = 32

var (
	ErrSessionNotFound       error = errors.New("Session not found")
	ErrSessionCreationFailed error = errors.New("Could not create session")

	NotLoggedInError = errors.New("Not logged in")
)

type SessionData struct {
	Username  string
	UserID    uuid.UUID
	CsrfToken []byte
	CreatedAt time.Time
}

type ManagerBackend interface {
	Set(key string, value *SessionData) error
	Get(key string) (*SessionData, error)
	Del(key string) error
}

func GenerateSessionKey() (string, error) {
	randomBytes, err := genbytes.GenRandomBytes(SESSION_KEY_LENGTH_BYTES)
	if err != nil {
		return "", ErrSessionCreationFailed
	}

	return fmt.Sprintf("%x", randomBytes), nil
}

func GenerateCSRFToken() ([]byte, error) {
	randomBytes, err := genbytes.GenRandomBytes(CSRF_TOKEN_LENGTH_BYTES)
	if err != nil {
		return nil, ErrSessionCreationFailed
	}

	return randomBytes, nil
}

type Manager struct {
	ManagerBackend
}

func (m *Manager) NewSession(username string, id uuid.UUID) (string, error) {
	key, err := GenerateSessionKey()
	if err != nil {
		return "", err
	}

	csrf, err := GenerateCSRFToken()
	if err != nil {
		return "", err
	}

	for {
		_, err := m.Get(key)
		if errors.Is(err, ErrSessionNotFound) {
			break
		} else if err != nil {
			return "", err
		}
		key, err = GenerateSessionKey()
		if err != nil {
			return "", err
		}
	}

	m.Set(key, &SessionData{
		Username:  username,
		UserID:    id,
		CsrfToken: csrf,
		CreatedAt: time.Now(),
	})

	return key, nil
}

func (m *Manager) GetSessionData(key string) (*SessionData, error) {
	return m.Get(key)
}

func (m *Manager) SetSessionData(key string, value *SessionData) error {
	_, err := m.Get(key)
	if err != nil {
		return err
	}
	return m.Set(key, value)
}

func (m *Manager) RefreshSession(key string) (string, error) {
	data, err := m.Get(key)
	if err != nil {
		return "", err
	}

	newKey, err := GenerateSessionKey()
	if err != nil {
		return "", err
	}

	for {
		_, err := m.Get(newKey)
		if errors.Is(err, ErrSessionNotFound) {
			break
		} else if err != nil {
			return "", err
		}
		newKey, err = GenerateSessionKey()
		if err != nil {
			return "", err
		}
	}

	m.Del(key)
	m.Set(newKey, data)

	return newKey, nil
}

func (m *Manager) EndSession(key string) error {
	return m.Del(key)
}
