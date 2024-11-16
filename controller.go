package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Fekinox/chrysalis-backend/internal/config"
	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
	"github.com/Fekinox/chrysalis-backend/internal/htmlrenderer"
	"github.com/Fekinox/chrysalis-backend/internal/models"
	"github.com/Fekinox/chrysalis-backend/internal/session"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChrysalisController struct {
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

func CreateController(cfg config.Config) (*ChrysalisController, error) {
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
	render.AddIncludes("templates/includes")
	render.AddTemplates("templates/templates")
	engine.HTMLRender = render

	return &ChrysalisController{
		cfg:    cfg,
		pool:   pool,
		router: engine,
		store:  store,

		sessionManager: session.NewMemorySessionManager(),
	}, nil
}

func (dc *ChrysalisController) Start(addr string) error {
	return dc.router.Run(addr)
}

func (dc *ChrysalisController) Router() *gin.Engine {
	return dc.router
}

func (dc *ChrysalisController) Close() error {
	dc.pool.Close()
	return nil
}

func (dc *ChrysalisController) IsUser(c *gin.Context, user string) bool {
	if user == "" {
		return false
	}
	data, ok := GetSessionData(c)
	if !ok {
		return false
	}
	return data.Username == user
}

func (dc *ChrysalisController) DummyHandler(c *gin.Context) {
	sessionKey, _ := GetSessionKey(c)
	sessionData, _ := GetSessionData(c)
	c.JSON(http.StatusOK, gin.H{
		"method":      c.Request.Method,
		"url":         c.Request.URL.RequestURI(),
		"sessionKey":  sessionKey,
		"sessionData": sessionData,
	})
}

func (dc *ChrysalisController) DummyTemplateHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "test.html.tmpl", nil)
}

func (dc *ChrysalisController) Login(c *gin.Context) {
	var loginSchema LoginSchema
	if err := c.ShouldBind(&loginSchema); err != nil {
		AbortError(c, http.StatusBadRequest, ErrBadRequest)
		return
	}

	// Retrieve user from database
	u, err := dc.store.GetUserByUsername(
		c.Request.Context(),
		loginSchema.Username,
	)
	if err != nil {
		AbortError(c, http.StatusBadRequest, ErrLoginFailed)
		return
	}

	// Compare password with hashed version
	ok, err := ComparePasswordAndHash(loginSchema.Password, u.Password)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	} else if !ok {
		AbortError(c, http.StatusForbidden, ErrLoginFailed)
		return
	}

	// Create session
	sessionKey, err := dc.sessionManager.NewSession(u.Username, u.ID)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
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
		AbortError(c, http.StatusBadRequest, ErrBadRequest)
		return
	}

	passHash, err := HashPassword(loginSchema.Password, DefaultParams())
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	userParams := db.CreateUserParams{
		Username: loginSchema.Username,
		Password: passHash,
	}

	u, err := dc.store.CreateUser(c.Request.Context(), userParams)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			AbortError(c, http.StatusConflict, ErrUserAlreadyExists)
		} else {
			AbortError(c, http.StatusInternalServerError, pgErr)
		}
		return
	} else if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	// Create initial user session
	sessionKey, err := dc.sessionManager.NewSession(u.Username, u.ID)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
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
		AbortError(c,
			http.StatusInternalServerError,
			errors.New("Session key not set"),
		)
		return
	}

	err := dc.sessionManager.EndSession(key)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
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
	_ = dc.store.BeginFunc(c.Request.Context(), func(s *db.Store) error {
		username := c.Param("username")
		if username == "" {
			err := errors.New("Must provide username")
			AbortError(c,
				http.StatusBadRequest,
				err,
			)
			return err
		}

		user, err := s.GetUserByUsername(c.Request.Context(), username)
		if err != nil {
			AbortError(c, http.StatusNotFound, errors.New("User not found"))
			return err
		}

		services, err := s.GetUserFormHeaders(c.Request.Context(), user.ID)
		if err != nil {
			AbortError(c, http.StatusNotFound, errors.New("Services not found"))
			return err
		}

		c.JSON(http.StatusAccepted, services)
		return nil
	})
}

func (dc *ChrysalisController) GetServiceBySlug(c *gin.Context) {
	params := models.ServiceFormParams{
		Username: c.Param("username"),
		Service:  c.Param("servicename"),
	}
	if params.Username == "" || params.Service == "" {
		AbortError(c,
			http.StatusBadRequest,
			errors.New("Must provide username and service name"),
		)
		return
	}

	form, err := models.GetServiceForm(
		c.Request.Context(),
		dc.store,
		params,
	)
	if err != nil {
		if errors.Is(err,
			errors.Join(
				models.ErrUserNotFound,
				models.ErrServiceNotFound,
				models.ErrFieldsNotFound,
			),
		) {
			AbortError(c, http.StatusNotFound, err)
			return
		}

		AbortError(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, form)
}

type NewServiceSpec struct {
	Title       string                `json:"title"`
	Slug        string                `json:"slug"`
	Description string                `json:"description"`
	Fields      []formfield.FormField `json:"fields"`
}

func (dc *ChrysalisController) CreateService(c *gin.Context) {
	username := c.Param("username")

	// Prevent creating a service if the username in the url does not match the
	// logged in user
	if !dc.IsUser(c, username) {
		AbortError(c, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	sessionData, _ := GetSessionData(c)

	var spec NewServiceSpec
	err := c.ShouldBindJSON(&spec)
	if err != nil {
		AbortError(c, http.StatusBadRequest, err)
		return
	}

	_, err = models.CreateServiceForm(
		c.Request.Context(),
		dc.store,
		models.CreateServiceVersionParams{
			CreatorID:   sessionData.UserID,
			ServiceSlug: spec.Slug,
			Title:       spec.Title,
			Description: spec.Description,
			Fields:      spec.Fields,
		},
	)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther,
		fmt.Sprintf("/api/users/%s/services/%s", username, spec.Slug))
}

type UpdateServiceSpec struct {
	Title       string                `json:"title"`
	Slug        string                `json:"slug"`
	Description string                `json:"description"`
	Fields      []formfield.FormField `json:"fields"`
}

func (dc *ChrysalisController) UpdateService(c *gin.Context) {
	username := c.Param("username")
	slug := c.Param("servicename")

	if slug == "" {
		AbortError(c, http.StatusBadRequest, errors.New("Missing slug"))
		return
	}

	// Prevent creating a service if the username in the url does not match the
	// logged in user
	if !dc.IsUser(c, username) {
		AbortError(c, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	sessionData, _ := GetSessionData(c)

	var spec UpdateServiceSpec
	err := c.ShouldBindJSON(&spec)
	if err != nil {
		AbortError(c, http.StatusBadRequest, err)
		return
	}

	_, err = models.UpdateServiceForm(
		c.Request.Context(),
		dc.store,
		models.CreateServiceVersionParams{
			CreatorID:   sessionData.UserID,
			ServiceSlug: slug,
			Title:       spec.Title,
			Description: spec.Description,
			Fields:      spec.Fields,
		},
	)
	if err != nil {
		if errors.Is(err, models.ErrUnchangedForm) {
			c.Redirect(http.StatusSeeOther,
				fmt.Sprintf("/api/users/%s/services/%s", username, spec.Slug))
			return
		} else {
			AbortError(c, http.StatusInternalServerError, err)
			return
		}
	}

	c.Redirect(http.StatusSeeOther,
		fmt.Sprintf("/api/users/%s/services/%s", username, spec.Slug))
}

func (dc *ChrysalisController) DeleteService(c *gin.Context) {
	username := c.Param("username")
	slug := c.Param("servicename")

	if !dc.IsUser(c, username) {
		AbortError(c, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	sessionData, _ := GetSessionData(c)

	if username == "" || slug == "" {
		AbortError(c,
			http.StatusBadRequest,
			errors.New("Must provide username and service name"),
		)
		return
	}

	err := dc.store.DeleteForm(c.Request.Context(), db.DeleteFormParams{
		Slug:      slug,
		CreatorID: sessionData.UserID,
	})
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (dc *ChrysalisController) GetUserServicesHTML(c *gin.Context) {
	_ = dc.store.BeginFunc(c.Request.Context(), func(s *db.Store) error {
		username := c.Param("username")
		if username == "" {
			err := errors.New("Must provide username")
			AbortError(c,
				http.StatusBadRequest,
				err,
			)
			return err
		}

		user, err := s.GetUserByUsername(c.Request.Context(), username)
		if err != nil {
			AbortError(c, http.StatusNotFound, errors.New("User not found"))
			return err
		}

		services, err := s.GetUserFormHeaders(c.Request.Context(), user.ID)
		if err != nil {
			AbortError(c, http.StatusNotFound, errors.New("Services not found"))
			return err
		}

		c.HTML(http.StatusOK, "userServices.html.tmpl", gin.H{
			"user":     user.Username,
			"services": services,
		})
		return nil
	})
}

func (dc *ChrysalisController) GetServiceDetail(c *gin.Context) {
	params := models.ServiceFormParams{
		Username: c.Param("username"),
		Service:  c.Param("servicename"),
	}
	if params.Username == "" || params.Service == "" {
		AbortError(c,
			http.StatusBadRequest,
			errors.New("Must provide username and service name"),
		)
		return
	}

	form, err := models.GetServiceForm(
		c.Request.Context(),
		dc.store,
		params,
	)
	if err != nil {
		if errors.Is(err,
			errors.Join(
				models.ErrUserNotFound,
				models.ErrServiceNotFound,
				models.ErrFieldsNotFound,
			),
		) {
			AbortError(c, http.StatusNotFound, err)
			return
		}

		AbortError(c, http.StatusNotFound, err)
		return
	}

	c.HTML(http.StatusOK, "serviceDetail.html.tmpl", form)
}

func (dc *ChrysalisController) ServiceCreator(c *gin.Context) {
	c.HTML(http.StatusOK, "serviceCreator.html.tmpl", nil)
}

func (dc *ChrysalisController) LoginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html.tmpl", nil)
}

func (dc *ChrysalisController) RegisterForm(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html.tmpl", nil)
}

func (dc *ChrysalisController) HandleLogin(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/app/login")
		return
	}
	username, ok := c.GetPostForm("username")
	if !ok {
		c.HTML(http.StatusOK, "login.html.tmpl", gin.H{
			"errors": "Username field missing",
		})
		return
	}
	password, ok := c.GetPostForm("password")
	if !ok {
		c.HTML(http.StatusOK, "login.html.tmpl", gin.H{
			"errors": "Password field missing",
		})
		return
	}

	// Retrieve user from database
	u, err := dc.store.GetUserByUsername(
		c.Request.Context(),
		username,
	)
	if err != nil {
		c.HTML(http.StatusOK, "login.html.tmpl", gin.H{
			"errors": "Invalid username or password",
		})
		return
	}

	// Compare password with hashed version
	ok, err = ComparePasswordAndHash(password, u.Password)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	} else if !ok {
		c.HTML(http.StatusOK, "login.html.tmpl", gin.H{
			"errors": "Invalid username or password",
		})
		return
	}

	// Create session
	sessionKey, err := dc.sessionManager.NewSession(u.Username, u.ID)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
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

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/app/%s/services", username))
}

func (dc *ChrysalisController) HandleRegister(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/app/register")
		return
	}
	username, ok := c.GetPostForm("username")
	if !ok {
		c.HTML(http.StatusOK, "register.html.tmpl", gin.H{
			"errors": "Username field missing",
		})
		return
	}
	password, ok := c.GetPostForm("password")
	if !ok {
		c.HTML(http.StatusOK, "register.html.tmpl", gin.H{
			"errors": "Password field missing",
		})
		return
	}

	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	passHash, err := HashPassword(password, DefaultParams())
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	userParams := db.CreateUserParams{
		Username: username,
		Password: passHash,
	}

	u, err := dc.store.CreateUser(c.Request.Context(), userParams)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			c.HTML(http.StatusOK, "register.html.tmpl", gin.H{
				"errors": ErrUserAlreadyExists.Error(),
			})
			return
		} else {
			AbortError(c, http.StatusInternalServerError, pgErr)
		}
		return
	} else if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	// Create initial user session
	sessionKey, err := dc.sessionManager.NewSession(u.Username, u.ID)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
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

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/app/%s/services", username))
}
