package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/config"
	"github.com/Fekinox/chrysalis-backend/internal/session"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const MAX_API_KEY_LENGTH = 64

var CSRF_SAFE_METHODS []string = []string{
	"GET",
	"HEAD",
	"OPTIONS",
	"TRACE",
}

var (
	MissingAPIKeyError         = errors.New("Could not find API key")
	InvalidAuthenticationError = errors.New("Invalid authentication")

	RateLimitExceededError = errors.New("Rate limit exceeded")

	TimeoutError = errors.New("Timeout")

	ErrNotLoggedIn   = errors.New("Not logged in")
	ErrForbidden     = errors.New("Forbidden")
	ErrCSRFViolation = errors.New("CSRF Violation- provide CSRF token")

	ErrMissingParam = errors.New("Param does not exist")
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
func ErrorHandler(cfg *config.Config, renderer ErrorRenderer) gin.HandlerFunc {
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

		// Only show detailed errors if either we are in dev mode or if the error is not
		// an internal error (error code >= 500)
		if cfg.Environment == "dev" || (c.Writer.Status() < 500 &&
			c.Writer.Status() != http.StatusForbidden) {
			c.JSON(-1, gin.H{"errors": errList})
		} else {
			c.JSON(-1, gin.H{"errors": http.StatusText(c.Writer.Status())})
		}
	}
}

type ErrorRenderer func(c *gin.Context, data any)

func JSONErrorRenderer(c *gin.Context, data any) {
	c.JSON(-1, data)
}

func HTMLErrorRenderer(c *gin.Context, data any) {
	c.HTML(-1, "errorPage.html.tmpl", data)
}

func RedirectTo(dests ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ContextRedirect(c, http.StatusSeeOther, dests...)
		c.Abort()
	}
}

// Redirects non-HTMX requests to the given URI.
func HTMXRedirect(dests ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("HX-Request") == "true" {
			c.Next()
		} else {
			ContextRedirect(c, http.StatusSeeOther, dests...)
			c.Abort()
		}
	}
}

// Redirect user to given URL, reusing path params if necessary.
func RedirectIfNotLoggedIn(sm *session.Manager, dests ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if data, ok := GetSessionData(c); ok && data.LoggedIn {
			c.Next()
		} else {
			ContextRedirect(c, http.StatusSeeOther, dests...)
			c.Abort()
		}
	}
}

func RedirectToLogin(sm *session.Manager) gin.HandlerFunc {
	return RedirectIfNotLoggedIn(sm, "/app/login")
}

// If the user has a valid session key cookie, then extract the session data and attach it to the
// request.
// If the user does not have a valid session key cookie, then create one.
func SessionKey(sm *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var ses session.SessionData
		if ses, err = extractSessionCookie(sm, c); err == nil {
			c.Set("sessionKey", ses.Key)
			c.Set("sessionData", ses)
			c.Next()
		} else if errors.Is(err, session.ErrSessionNotFound) {
			key, data, err := sm.NewSession()
			if err != nil {
				AbortError(c, http.StatusInternalServerError, err)
				return
			}

			c.SetCookie(
				"chrysalis-session-key",
				key,
				60*60*24,
				"/",
				"localhost",
				false,
				true,
			)
			c.SetSameSite(http.SameSiteLaxMode)

			c.Set("sessionKey", key)
			c.Set("sessionData", data)
			c.Next()
		} else {
			AbortError(c, http.StatusInternalServerError, err)
		}
	}
}

func extractSessionCookie(sm *session.Manager, c *gin.Context) (session.SessionData, error) {
	key, err := c.Request.Cookie("chrysalis-session-key")
	if err != nil {
		return session.SessionData{}, err
	}
	data, err := sm.GetSessionData(key.Value)
	if err != nil {
		return session.SessionData{}, err
	}

	return data, err
}

func contextGetter[T any](key any) func(c *gin.Context) (T, bool) {
	return func(c *gin.Context) (T, bool) {
		v, ok := c.Value(key).(T)
		return v, ok
	}
}

var (
	GetSessionData = contextGetter[session.SessionData]("sessionData")
	GetSessionKey  = contextGetter[string]("sessionKey")
)

func HasSessionKey(sm *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := GetSessionData(c)
		if !ok {
			AbortError(c, http.StatusForbidden, ErrNotLoggedIn)
			return
		}

		c.Next()
	}
}

func CSRFTokenExists(r *http.Request) bool {
	return r.Header.Get("X-CSRF-Protection") != ""
}

// Checks for the CSRF token in the `X-Csrf-Token` header.
func GetCSRFTokenInHeaders(r *http.Request) ([]byte, bool) {
	var token string
	if r.Header.Get("X-CSRF-Protection") == "" {
		return nil, false
	}
	token = r.Header.Get("X-CSRF-Token")
	if token == "" {
		return nil, false
	}
	decodedHeader, err := hex.DecodeString(token)
	if err != nil {
		return nil, false
	}
	return decodedHeader, true
}

// Checks for the CSRF token in the form data, often sent with simple forms.
func GetCSRFTokenInFormData(r *http.Request) ([]byte, bool) {
	err := r.ParseForm()
	if err != nil {
		return nil, false
	}

	token := r.FormValue("csrf-token")
	if token == "" {
		return nil, false
	}
	decodedHeader, err := hex.DecodeString(token)
	if err != nil {
		return nil, false
	}
	return decodedHeader, true
}

// Compares the CSRF token sent in the request headers
func CsrfProtect(sm *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if slices.Contains(CSRF_SAFE_METHODS, c.Request.Method) {
			c.Next()
			return
		}
		session, ok := GetSessionData(c)
		if !ok {
			AbortError(c, http.StatusForbidden, ErrForbidden)
			return
		}

		token, ok := GetCSRFTokenInHeaders(c.Request)
		if !ok {
			token, ok = GetCSRFTokenInFormData(c.Request)
			if !ok {
				AbortError(c, http.StatusForbidden, ErrCSRFViolation)
				return
			}
		}

		if subtle.ConstantTimeCompare(token, session.CsrfToken) == 0 {
			AbortError(c, http.StatusForbidden, ErrCSRFViolation)
			return
		}
		c.Next()
	}
}

// Checks the request's header for a CSRF protection flag. Merely checking for the existence of
// this header provides CORS protection for fetch and XHR requests: any request with non-simple
// headers is subject to a CORS preflight check if it originated from a web browser. Full CSRF
// protection is necessary if a site sends simple requests or uses form tags.
func HasCSRFHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if slices.Contains(CSRF_SAFE_METHODS, c.Request.Method) || CSRFTokenExists(c.Request) {
			c.Next()
		} else {
			AbortError(c, http.StatusForbidden, ErrCSRFViolation)
		}
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

// Given the gin context which has already parsed path parameters, attempt to parse the given URL.
func ParseURL(c *gin.Context, path string) (string, error) {
	tokens := strings.Split(path, "/")
	for i := 0; i < len(tokens); i++ {
		p, found := strings.CutPrefix(tokens[i], ":")
		if found {
			param := c.Param(p)
			if param == "" {
				return "", fmt.Errorf("%w: %s", ErrMissingParam, p)
			}
			tokens[i] = param
		}
	}

	return strings.Join(tokens, "/"), nil
}

func ParseFallbackURLs(c *gin.Context, paths ...string) (string, error) {
	for _, p := range paths {
		res, err := ParseURL(c, p)
		if err != nil {
			continue
		}
		return res, nil
	}
	return "", errors.New("Could not resolve fallback URLs")
}

func ContextRedirect(c *gin.Context, code int, paths ...string) {
	dest, err := ParseFallbackURLs(c, paths...)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	} else {
		c.Redirect(code, dest)
	}
}
