package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/config"
	"github.com/Fekinox/chrysalis-backend/internal/session"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const MAX_API_KEY_LENGTH = 64

var (
	MissingAPIKeyError         = errors.New("Could not find API key")
	InvalidAuthenticationError = errors.New("Invalid authentication")

	RateLimitExceededError = errors.New("Rate limit exceeded")

	TimeoutError = errors.New("Timeout")

	NotLoggedInError = errors.New("Not logged in")
)

// Extracts the token from the given header in the request.
func extractToken(c *gin.Context, header string) (string, error) {
	authString := c.GetHeader(header)
	tokens := strings.SplitN(authString, " ", 2)
	if len(tokens) < 2 {
		return "", errors.New("Invalid token")
	}

	token := strings.TrimSpace(tokens[1])
	if len(token) > MAX_API_KEY_LENGTH {
		return "", errors.New("Invalid token")
	}

	return strings.TrimSpace(token), nil
}

// Verifies the raw API key by hashing it with the SHA-256 algorithm and
// doing a constant-time comparison against the server's API key. A
// constant-time comparison is necessary to prevent timing attacks: an
// adversary could guess the key by measuring how long it takes for a naive
// equality algorithm to return.
func verifyApiKey(cfg *config.Config, s string) bool {
	hash := sha256.Sum256([]byte(s))
	key := hash[:]

	return subtle.ConstantTimeCompare(cfg.DecodedAPIKey, key) == 1
}

// Rejects all requests that do not have a valid API key.
func ApiKeyMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		key, err := extractToken(c, "Authorization")
		if err != nil {
			AbortError(c, http.StatusUnauthorized, MissingAPIKeyError)
			return
		}

		if !verifyApiKey(cfg, key) {
			AbortError(c,
				http.StatusUnauthorized,
				InvalidAuthenticationError,
			)
			return
		}

		c.Next()
	}
}

// Applies a rate limiter to the http handler. At most r events will be sent to
// the handler per second, and it also permits bursts of up to b events.
func RateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	limiter := rate.NewLimiter(r, b)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			AbortError(c, http.StatusTooManyRequests, RateLimitExceededError)
			return
		} else {
			c.Next()
		}
	}
}

// Adds an artificial delay to the event handler.
func Delay(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		time.Sleep(timeout)
		c.Next()
	}
}

// Handles all errors raised by event handlers and middleware.
func ErrorHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		errList := make([]string, len(c.Errors))

		for i, ginErr := range c.Errors {
			fmt.Println(ginErr)
			errList[i] = ginErr.Error()
		}

		// Only show errors if either we are in dev mode or if the error is not
		// an internal error (error code >= 500)
		if cfg.Environment == "dev" || c.Writer.Status() < 500 {
			c.JSON(-1, gin.H{"errors": errList})
		} else {
			c.JSON(-1, gin.H{"errors": http.StatusText(c.Writer.Status())})
		}
	}
}

// Reads the session key cookie from the client and adds it to the context
func SessionKey(sm session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey, err := c.Request.Cookie("chrysalis-session-key")
		if err != nil {
			c.Next()
			return
		}
		sessionData, err := sm.GetSessionData(sessionKey.Value)
		if err != nil {
			c.Next()
			return
		}

		c.Set("sessionKey", sessionKey.Value)
		c.Set("sessionData", sessionData)

		c.Next()
	}
}

func contextGetter[T any](key any) func(c *gin.Context) (T, bool) {
	return func(c *gin.Context) (T, bool) {
		v, ok := c.Value(key).(T)
		return v, ok
	}
}

var (
	GetSessionData = contextGetter[*session.SessionData]("sessionData")
	GetSessionKey  = contextGetter[string]("sessionKey")
)

func HasSessionKey(sm session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := GetSessionData(c)
		if !ok {
			AbortError(c, http.StatusForbidden, NotLoggedInError)
			return
		}

		c.Next()
	}
}

func AbortStatus(c *gin.Context, code int) {
	c.Status(code)
	c.Abort()
}

func AbortError(c *gin.Context, code int, err error) {
	c.Status(code)
	c.Error(err)
	c.Abort()
}
