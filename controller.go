package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Fekinox/chrysalis-backend/internal/config"
	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
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

	fmt.Println(cfg.GetDBUrl())

	pool, err := pgxpool.New(context.Background(), cfg.GetDBUrl())
	if err != nil {
		return nil, err
	}

	engine.LoadHTMLGlob("templates/*")

	return &ChrysalisController{
		cfg:    cfg,
		pool:   pool,
		router: engine,

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

// Creates a new db wrapper from the connection pool.
func (dc *ChrysalisController) Store(ctx context.Context) (*db.Store, error) {
	conn, err := dc.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return db.NewStore(conn), nil
}

// Executes the given function with a fresh DB connection from the pool.
func (dc *ChrysalisController) StoreFunc(ctx context.Context, fn func(s *db.Store) error) error {
	return dc.pool.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		return fn(db.NewStore(c))
	})
}

// Executes the given function with a fresh DB connection from the pool, wrapped in a transaction.
func (dc *ChrysalisController) StoreFuncTx(ctx context.Context, fn func(s *db.Store) error) error {
	return dc.pool.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		store := db.NewStore(c)
		return store.BeginFunc(ctx, func(s *db.Store) error {
			return fn(s)
		})
	})
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
	c.HTML(http.StatusOK, "example.html.tmpl", nil)
}

func (dc *ChrysalisController) Login(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		var loginSchema LoginSchema
		if err := c.ShouldBind(&loginSchema); err != nil {
			AbortError(c, http.StatusBadRequest, ErrBadRequest)
			return err
		}

		// Retrieve user from database
		u, err := s.GetUserByUsername(
			c.Request.Context(),
			loginSchema.Username,
		)
		if err != nil {
			return ErrLoginFailed
		}

		// Compare password with hashed version
		ok, err := ComparePasswordAndHash(loginSchema.Password, u.Password)
		if err != nil {
			return err
		} else if !ok {
			return ErrLoginFailed
		}

		// Create session
		sessionKey, err := dc.sessionManager.NewSession(u.Username, u.ID)
		if err != nil {
			return err
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

		return nil
	})

	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, ErrLoginFailed) {
			code = http.StatusForbidden
		} else if errors.Is(err, ErrBadRequest) {
			code = http.StatusBadRequest
		} else {
			AbortError(c, http.StatusInternalServerError, err)
		}
		AbortError(c, code, err)
	}
}

func (dc *ChrysalisController) Register(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		var loginSchema LoginSchema
		if err := c.ShouldBind(&loginSchema); err != nil {
			return ErrBadRequest
		}

		passHash, err := HashPassword(loginSchema.Password, DefaultParams())
		if err != nil {
			return err
		}

		userParams := db.CreateUserParams{
			Username: loginSchema.Username,
			Password: passHash,
		}

		u, err := s.CreateUser(c.Request.Context(), userParams)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrUserAlreadyExists
			} else {
				return pgErr
			}
		} else if err != nil {
			return err
		}

		// Create initial user session
		sessionKey, err := dc.sessionManager.NewSession(u.Username, u.ID)
		if err != nil {
			return err
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

		return nil
	})

	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, ErrBadRequest) {
			code = http.StatusBadRequest
		} else if errors.Is(err, ErrUserAlreadyExists) {
			code = http.StatusConflict
		}
		AbortError(c, code, err)
	}
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
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		return s.BeginFunc(c.Request.Context(), func(s *db.Store) error {
			username := c.Param("username")
			if username == "" {
				return errors.New("Must provide username")
			}

			user, err := s.GetUserByUsername(c.Request.Context(), username)
			if err != nil {
				return errors.New("User not found")
			}

			services, err := s.GetUserFormHeaders(c.Request.Context(), user.ID)
			if err != nil {
				return errors.New("Services not found")
			}

			c.JSON(http.StatusAccepted, services)
			return nil
		})
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

func (dc *ChrysalisController) GetServiceBySlug(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		params := models.ServiceFormParams{
			Username: c.Param("username"),
			Service:  c.Param("servicename"),
		}
		if params.Username == "" || params.Service == "" {
			return errors.New("Must provide username and service name")
		}

		form, err := models.GetServiceForm(
			c.Request.Context(),
			s,
			params,
		)
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, form)
		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

type NewServiceSpec struct {
	Title       string                `json:"title"`
	Slug        string                `json:"slug"`
	Description string                `json:"description"`
	Fields      []formfield.FormField `json:"fields"`
}

func (dc *ChrysalisController) CreateService(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		username := c.Param("username")

		// Prevent creating a service if the username in the url does not match the
		// logged in user
		if !dc.IsUser(c, username) {
			return errors.New("Unauthorized")
		}
		sessionData, _ := GetSessionData(c)

		var spec NewServiceSpec
		err := c.ShouldBindJSON(&spec)
		if err != nil {
			return err
		}

		_, err = models.CreateServiceForm(
			c.Request.Context(),
			s,
			models.CreateServiceVersionParams{
				CreatorID:   sessionData.UserID,
				ServiceSlug: spec.Slug,
				Title:       spec.Title,
				Description: spec.Description,
				Fields:      spec.Fields,
			},
		)
		if err != nil {
			return err
		}

		c.Redirect(http.StatusSeeOther,
			fmt.Sprintf("/api/users/%s/services/%s", username, spec.Slug))
		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

type UpdateServiceSpec struct {
	Title       string                `json:"title"`
	Slug        string                `json:"slug"`
	Description string                `json:"description"`
	Fields      []formfield.FormField `json:"fields"`
}

func (dc *ChrysalisController) UpdateService(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		username := c.Param("username")
		slug := c.Param("servicename")

		if slug == "" {
			return errors.New("Missing slug")
		}

		// Prevent creating a service if the username in the url does not match the
		// logged in user
		if !dc.IsUser(c, username) {
			return errors.New("Unauthorized")
		}
		sessionData, _ := GetSessionData(c)

		var spec UpdateServiceSpec
		err := c.ShouldBindJSON(&spec)
		if err != nil {
			return err
		}

		_, err = models.UpdateServiceForm(
			c.Request.Context(),
			s,
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
				return nil
			} else {
				return err
			}
		}

		c.Redirect(http.StatusSeeOther,
			fmt.Sprintf("/api/users/%s/services/%s", username, spec.Slug))

		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

func (dc *ChrysalisController) DeleteService(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		username := c.Param("username")
		slug := c.Param("servicename")

		if !dc.IsUser(c, username) {
			return errors.New("Unauthorized")
		}
		sessionData, _ := GetSessionData(c)

		if username == "" || slug == "" {
			return errors.New("Must provide username and service name")
		}

		err := s.DeleteForm(c.Request.Context(), db.DeleteFormParams{
			Slug:      slug,
			CreatorID: sessionData.UserID,
		})
		if err != nil {
			return err
		}

		c.Status(http.StatusNoContent)

		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

func (dc *ChrysalisController) GetUserServicesHTML(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		return s.BeginFunc(c.Request.Context(), func(s *db.Store) error {
			username := c.Param("username")
			if username == "" {
				err := errors.New("Must provide username")
				return err
			}

			user, err := s.GetUserByUsername(c.Request.Context(), username)
			if err != nil {
				return errors.New("User not found")
			}

			services, err := s.GetUserFormHeaders(c.Request.Context(), user.ID)
			if err != nil {
				return errors.New("Services not found")
			}

			c.HTML(http.StatusOK, "userServices.html.tmpl", gin.H{
				"user":     user.Username,
				"services": services,
			})
			return nil
		})
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

func (dc *ChrysalisController) GetServiceDetail(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		params := models.ServiceFormParams{
			Username: c.Param("username"),
			Service:  c.Param("servicename"),
		}
		if params.Username == "" || params.Service == "" {
			return errors.New("Must provide username and service name")
		}

		form, err := models.GetServiceForm(
			c.Request.Context(),
			s,
			params,
		)
		if err != nil {
			return err
		}

		c.HTML(http.StatusOK, "serviceDetail.html.tmpl", form)
		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}
