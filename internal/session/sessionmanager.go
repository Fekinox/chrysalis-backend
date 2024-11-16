package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/genbytes"
	"github.com/google/uuid"
)

const SESSION_KEY_LENGTH_BYTES = 16

var (
	ErrSessionNotFound       error = errors.New("Session not found")
	ErrSessionCreationFailed error = errors.New("Could not create session")

	NotLoggedInError = errors.New("Not logged in")
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

func GenerateSessionKey() (string, error) {
	randomBytes, err := genbytes.GenRandomBytes(SESSION_KEY_LENGTH_BYTES)
	if err != nil {
		return "", ErrSessionCreationFailed
	}

	return fmt.Sprintf("%x", randomBytes), nil
}
