package session

import (
	"errors"

	"github.com/google/uuid"
)

var (
	SessionNotFound error = errors.New("Session not found")
)

type SessionKey string

type SessionData struct {
	Username string
	UserID   uuid.UUID
}

type Manager interface {
	// Initializes a new user session with the given username and id.
	NewSession(username string, id uuid.UUID) SessionKey
	// Gets the session data owned by the given session key
	GetSessionData(key SessionKey) (*SessionData, error)
	// Sets the session data owned by the given session key
	SetSessionData(key SessionKey, data *SessionData) error
	// Refreshes the session; invalidates the previous key and generates a new
	// one pointing to the same session data
	RefreshSession(key SessionKey) (SessionKey, error)
	// Ends the session and invalidates the key.
	EndSession(key SessionKey) error
}
