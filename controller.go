package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/config"
	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/htmlrenderer"
	"github.com/Fekinox/chrysalis-backend/internal/session"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChrysalisServer struct {
	cfg    config.Config
	router *gin.Engine
	pool   *pgxpool.Pool
	store  *db.Store

	sessionManager session.Manager
}

type LoginSchema struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

var (
	ErrBadRequest        = errors.New("Bad request")
	ErrNotFound          = errors.New("Resource not found")
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrLoginFailed       = errors.New("Login failed")
)

func CreateController(cfg config.Config) (*ChrysalisServer, error) {
	var engine *gin.Engine
	if cfg.Environment == "test" || cfg.Environment == "release" {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
	} else {
		engine = gin.Default()
	}

	fmt.Println(cfg.GetDBUrl())

	pool, err := pgxpool.New(context.Background(), cfg.GetDBUrl())
	if err != nil {
		return nil, err
	}

	store := db.NewStore(pool)

	var render htmlrenderer.Renderer
	if gin.IsDebugging() {
		render = htmlrenderer.NewDebug()
	} else {
		render = htmlrenderer.New()
	}
	render.Funcs(template.FuncMap{
		"timeFormatUnix": func(t time.Time) string {
			return t.Format(time.UnixDate)
		},
		"contains": func(s string, strings []string) bool {
			return slices.Contains(strings, s)
		},
		"statuses": func() []string {
			return []string{
				"pending",
				"approved",
				"in progress",
				"delayed",
				"complete",
				"cancelled",
			}
		},
		"hyphenize": func(s string) string {
			return strings.ReplaceAll(s, " ", "-")
		},
	})
	render.AddIncludes("templates/includes")
	render.AddTemplates("templates/templates")
	engine.HTMLRender = render

	return &ChrysalisServer{
		cfg:    cfg,
		pool:   pool,
		router: engine,
		store:  store,

		sessionManager: session.NewMemorySessionManager(),
	}, nil
}

func (cs *ChrysalisServer) Start(addr string) error {
	return cs.router.Run(addr)
}

func (cs *ChrysalisServer) Router() *gin.Engine {
	return cs.router
}

func (cs *ChrysalisServer) Close() error {
	cs.pool.Close()
	return nil
}

func DummyHandler(c *gin.Context) {
	sessionKey, _ := GetSessionKey(c)
	sessionData, _ := GetSessionData(c)
	c.JSON(http.StatusOK, gin.H{
		"method":      c.Request.Method,
		"url":         c.Request.URL.RequestURI(),
		"sessionKey":  sessionKey,
		"sessionData": sessionData,
	})
}

func DummyTemplateHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "test.html.tmpl", nil)
}

func (cs *ChrysalisServer) Healthcheck(c *gin.Context) {
	c.HTML(http.StatusOK, "healthcheck.html.tmpl", nil)
}

func (cs *ChrysalisServer) HealthcheckInner(c *gin.Context) {
	stat := cs.pool.Stat()
	c.HTML(http.StatusOK, "healthcheckInner.html.tmpl", gin.H{
		"localTime":     time.Now(),
		"acquiredCount": stat.AcquireCount(),
		"acquiredConns": stat.AcquiredConns(),
		"totalConns":    stat.TotalConns(),
		"maxConns":      stat.MaxConns(),
	})
}

func (cs *ChrysalisServer) HealthcheckObjectStats(c *gin.Context) {
	stats, err := cs.store.GetChrysalisStats(c.Request.Context())
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "healthcheckObjects.html.tmpl", gin.H{
		"numUsers": stats.CountUsers,
		"numForms": stats.NumForms,
		"numTasks": stats.NumTasks,
	})
}
