package session

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/genbytes"
	"github.com/gin-gonic/gin"
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

type Session struct {
	manager Manager
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

func (s *Session) New(
	w http.ResponseWriter,
	username string, id uuid.UUID,
) (string, error) {
	sessionKey, err := s.manager.NewSession(username, id)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "chrysalis-session-key",
		Value:    sessionKey,
		MaxAge:   60 * 60 * 24,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	})
	return sessionKey, err
}

func (s *Session) Get(
	r *http.Request,
) (*SessionData, error) {
	sessionKey, err := r.Cookie("chrysalis-session-key")
	if err != nil {
		return nil, err
	}

	return s.manager.GetSessionData(sessionKey.Value)
}

func (s *Session) AddSessionData() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey, err := c.Request.Cookie("chrysalis-session-key")
		if err != nil {
			c.Next()
			return
		}

		sessionData, err := s.manager.GetSessionData(sessionKey.Value)
		if err != nil {
			c.Next()
			return
		}

		c.Set("sessionKey", sessionKey.Value)
		c.Set("sessionData", sessionData)

		c.Next()
	}
}

func (s *Session) HasSessionData() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := s.Get(c.Request)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Error(NotLoggedInError)
			c.Abort()
			return
		}

		c.Next()
	}
}
