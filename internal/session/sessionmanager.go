package session

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const SESSION_KEY_LENGTH_BYTES = 16

var (
	SessionNotFound error = errors.New("Session not found")
)

type SessionData struct {
	Username  string
	UserID    uuid.UUID
	CreatedAt time.Time
}

type Manager interface {
	// Initializes a new user session with the given username and id.
	NewSession(username string, id uuid.UUID) (string, error)
	// Gets the session data owned by the given session key
	GetSessionData(key string) (*SessionData, error)
	// Sets the session data owned by the given session key
	SetSessionData(key string, data *SessionData) error
	// Refreshes the session; invalidates the previous key and generates a new
	// one pointing to the same session data
	RefreshSession(key string) (string, error)
	// Ends the session and invalidates the key.
	EndSession(key string) error
}

func genRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateSessionKey() (string, error) {
	randomBytes, err := genRandomBytes(SESSION_KEY_LENGTH_BYTES)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", randomBytes), nil
}
