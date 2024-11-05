package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Fekinox/chrysalis-backend/internal/config"
	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
	"github.com/Fekinox/chrysalis-backend/internal/session"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ChrysalisController struct {
	db     *db.Queries
	cfg    config.Config
	router *gin.Engine
	conn   *pgx.Conn

	sessionManager session.Manager
}

type LoginSchema struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

var (
	ErrBadRequest = errors.New("Bad request")
	ErrNotFound   = func(name string) error {
		return errors.New(fmt.Sprintf("Not found: %s", name))
	}
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrLoginFailed       = errors.New("Login failed")
)

func CreateController(cfg config.Config) (*ChrysalisController, error) {
	var engine *gin.Engine
	if cfg.Environment == "test" || cfg.Environment == "release" {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
	} else {
		engine = gin.Default()
	}

	conn, err := pgx.Connect(context.Background(), cfg.GetDBUrl())
	if err != nil {
		return nil, err
	}

	q := db.New(conn)

	return &ChrysalisController{
		db:     q,
		cfg:    cfg,
		conn:   conn,
		router: engine,

		sessionManager: session.NewMemorySessionManager(),
	}, nil
}

func (dc *ChrysalisController) MountHandlers() {
	api := dc.router.Group("/api")
	api.Use(ErrorHandler(&dc.cfg))

	auth := api.Group("/auth")
	auth.POST("/login", dc.Login)
	auth.POST("/register", dc.Register)
	auth.POST("/logout", SessionKey(dc.sessionManager), dc.Logout)

	users := api.Group("/users")
	// Get all services a user owns
	users.GET("/:username/services", dc.DummyHandler)
	users.GET("/:username/services/:servicename", dc.GetServiceBySlug)

	users.POST("/:username/services",
		SessionKey(dc.sessionManager),
		dc.CreateService)
	users.PUT("/:username/services/:servicename", dc.DummyHandler)
	users.DELETE("/:username/services/:servicename", dc.DummyHandler)

	// Get outbound and inbound tasks for a user
	users.GET("/:username/outbound-tasks", dc.DummyHandler)
	users.GET("/:username/inbound-tasks", dc.DummyHandler)

	// Tasks associated with a particular service
	users.GET("/:username/services/:servicename/tasks", dc.DummyHandler)
	users.GET(
		"/:username/services/:servicename/tasks/:taskslug",
		dc.DummyHandler,
	)
	users.POST("/:username/services/:servicename/tasks", dc.DummyHandler)
	users.PUT(
		"/:username/services/:servicename/tasks/:taskslug",
		dc.DummyHandler,
	)
}

func (dc *ChrysalisController) Start(addr string) error {
	return dc.router.Run(addr)
}

func (dc *ChrysalisController) Router() *gin.Engine {
	return dc.router
}

func (dc *ChrysalisController) Close() error {
	return dc.conn.Close(context.Background())
}

func (dc *ChrysalisController) DummyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"method":      c.Request.Method,
		"url":         c.Request.URL.RequestURI(),
		"sessionKey":  c.Value("sessionKey"),
		"sessionData": c.Value("sessionData"),
	})
}

func (dc *ChrysalisController) Login(c *gin.Context) {
	var loginSchema LoginSchema
	if err := c.ShouldBind(&loginSchema); err != nil {
		c.AbortWithError(http.StatusBadRequest, ErrBadRequest)
		return
	}

	// Retrieve user from database
	u, err := dc.db.GetUserByUsername(
		c.Request.Context(),
		loginSchema.Username,
	)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, ErrLoginFailed)
		return
	}

	// Compare password with hashed version
	ok, err := ComparePasswordAndHash(loginSchema.Password, u.Password)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if !ok {
		c.AbortWithError(http.StatusForbidden, ErrLoginFailed)
		return
	}

	// Create session
	sessionKey, err := dc.sessionManager.NewSession(u.Username, u.ID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.SetCookie(
		"chrysalis-session-key",
		string(sessionKey),
		60*60*24,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"session_key": sessionKey,
		"username":    u.Username,
		"id":          u.ID,
	})
}

func (dc *ChrysalisController) Register(c *gin.Context) {
	var loginSchema LoginSchema
	if err := c.ShouldBind(&loginSchema); err != nil {
		c.AbortWithError(http.StatusBadRequest, ErrBadRequest)
		return
	}

	passHash, err := HashPassword(loginSchema.Password, DefaultParams())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	userParams := db.CreateUserParams{
		Username: loginSchema.Username,
		Password: passHash,
	}

	u, err := dc.db.CreateUser(c.Request.Context(), userParams)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			c.AbortWithError(http.StatusConflict, ErrUserAlreadyExists)
		} else {
			c.AbortWithError(http.StatusInternalServerError, pgErr)
		}
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Create initial user session
	sessionKey, err := dc.sessionManager.NewSession(u.Username, u.ID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.SetCookie(
		"chrysalis-session-key",
		sessionKey,
		60*60*24,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(http.StatusCreated, gin.H{
		"session_key": sessionKey,
		"username":    u.Username,
		"id":          u.ID,
	})
}

func (dc *ChrysalisController) Logout(c *gin.Context) {
	key, ok := c.Value("sessionKey").(string)
	if !ok {
		c.AbortWithError(
			http.StatusInternalServerError,
			errors.New("Session key not set"),
		)
		return
	}

	err := dc.sessionManager.EndSession(key)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.SetCookie(
		"chrysalis-session-key",
		"",
		0,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

func (dc *ChrysalisController) GetUserServices(c *gin.Context) {
	c.AbortWithError(
		http.StatusServiceUnavailable,
		errors.New("Unimplemented"),
	)
}

func (dc *ChrysalisController) GetServiceBySlug(c *gin.Context) {
	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	qtx := dc.db.WithTx(tx)

	username := c.Param("username")
	serviceSlug := c.Param("servicename")
	if username == "" || serviceSlug == "" {
		c.AbortWithError(
			http.StatusBadRequest,
			errors.New("Must provide username and service name"),
		)
		return
	}

	user, err := qtx.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, errors.New("User not found"))
	}

	params := db.GetCurrentFormVersionBySlugParams{
		Slug:      serviceSlug,
		CreatorID: user.ID,
	}

	service, err := qtx.GetCurrentFormVersionBySlug(c.Request.Context(), params)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, errors.New("Service not found"))
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, service)
}

type NewServiceSpec struct {
	Title       string                `json:"title"`
	Slug        string                `json:"slug"`
	Description string                `json:"description"`
	Fields      []formfield.FormField `json:"fields"`
}

func (dc *ChrysalisController) CreateService(c *gin.Context) {
	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c.Request.Context())

	_ = dc.db.WithTx(tx)

	username := c.Param("username")
	if !dc.IsUser(c, username) {
		c.AbortWithError(http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	var spec NewServiceSpec
	err = c.BindJSON(&spec)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, spec)
}

func (dc *ChrysalisController) UpdateService(c *gin.Context) {
	c.AbortWithError(
		http.StatusServiceUnavailable,
		errors.New("Unimplemented"),
	)
}

func (dc *ChrysalisController) DeleteService(c *gin.Context) {
	c.AbortWithError(
		http.StatusServiceUnavailable,
		errors.New("Unimplemented"),
	)
}

func (dc *ChrysalisController) IsUser(c *gin.Context, user string) bool {
	if user == "" || c.Value("sessionKey") == "" {
		return false
	}
	data, ok := c.Value("sessionData").(*session.SessionData)
	if !ok {
		return false
	}
	return data.Username == user
}
