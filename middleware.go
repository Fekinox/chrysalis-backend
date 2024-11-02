package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const MAX_API_KEY_LENGTH = 64

var (
	MissingAPIKeyError         = errors.New("Could not find API key")
	InvalidAuthenticationError = errors.New("Invalid authentication")
	RateLimitExceededError     = errors.New("Rate limit exceeded")
	TimeoutError               = errors.New("Timeout")
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
func verifyApiKey(cfg *Config, s string) bool {
	hash := sha256.Sum256([]byte(s))
	key := hash[:]

	return subtle.ConstantTimeCompare(cfg.DecodedAPIKey, key) == 1
}

// Rejects all requests that do not have a valid API key.
func ApiKeyMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		key, err := extractToken(c, "Authorization")
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, MissingAPIKeyError)
			return
		}

		if !verifyApiKey(cfg, key) {
			c.AbortWithError(
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
			c.AbortWithError(http.StatusTooManyRequests, RateLimitExceededError)
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
func ErrorHandler(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		errList := make([]string, len(c.Errors))

		for i, ginErr := range c.Errors {
			// fmt.Println(ginErr)
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
