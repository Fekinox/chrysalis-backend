package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type MainController struct {
	con *ChrysalisServer
}

func (mc *MainController) MountTo(path string, app gin.IRouter) {
	app.Use(ErrorHandler(&mc.con.cfg, HTMLErrorRenderer))
	app.Use(SessionKey(mc.con.sessionManager))
	app.GET("/helloworld", DummyTemplateHandler)
	app.GET("/:username/services", mc.GetUserServices)

	app.GET("/:username/services/:servicename/dashboard", mc.ServiceDashboard)
	app.GET(
		"/:username/services/:servicename/dashboard/tabs/:status",
		HTMXRedirect("/app/:username/services/:servicename"),
		mc.ServiceDashboardTab,
	)
	app.PUT(
		"/:username/services/:servicename/tasks/:taskname",
		HasSessionKey(mc.con.sessionManager),
		mc.UpdateTask,
	)

	app.GET(
		"/:username/services/:servicename/form",
		RedirectToLogin(mc.con.sessionManager),
		mc.ServiceDetail,
	)
	app.GET("/:username/services/:servicename/edit",
		RedirectToLogin(mc.con.sessionManager),
		mc.ServiceDetail,
	)

	app.GET("/new-service",
		RedirectToLogin(mc.con.sessionManager),
		mc.ServiceCreator,
	)
	app.POST("/new-service",
		HasSessionKey(mc.con.sessionManager),
		mc.CreateNewService,
	)
	app.GET("/:username/services/:servicename/tasks/:taskname", DummyTemplateHandler)

	app.GET("/login", mc.LoginForm)
	app.POST("/login", mc.HandleLogin)

	app.GET("/register", mc.RegisterForm)
	app.POST("/register", mc.HandleRegister)

	app.GET("/dashboard", RedirectToLogin(mc.con.sessionManager), mc.UserDashboard)
}

func NewMainController(c *ChrysalisServer) (*MainController, error) {
	mc := &MainController{
		con: c,
	}

	c.Mount("/app", mc)

	return mc, nil
}

func (dc *MainController) GetUserServices(c *gin.Context) {
	_ = dc.con.store.BeginFunc(c.Request.Context(), func(s *db.Store) error {
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

func (mc *MainController) UserDashboard(c *gin.Context) {
	sessionData, _ := GetSessionData(c)

	services, err := mc.con.store.GetUserFormHeaders(c.Request.Context(), sessionData.UserID)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "userDashboard.html.tmpl", gin.H{
		"session":  sessionData,
		"services": services,
	})
}

func (mc *MainController) ServiceDashboard(c *gin.Context) {
	sessionData, _ := GetSessionData(c)

	var form *models.ServiceForm
	var taskHeaders []*db.GetServiceTasksBySlugRow

	err := mc.con.store.BeginFunc(c.Request.Context(), func(s *db.Store) error {
		var err error
		form, err = models.GetServiceForm(c.Request.Context(), s, models.ServiceFormParams{
			Username: c.Param("username"),
			Service:  c.Param("servicename"),
		})
		if err != nil {
			return err
		}

		taskHeaders, err = s.GetServiceTasksBySlug(
			c.Request.Context(),
			db.GetServiceTasksBySlugParams{
				CreatorUsername: c.Param("username"),
				FormSlug:        c.Param("servicename"),
			},
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "serviceDashboard.html.tmpl", gin.H{
		"session": sessionData,
		"service": form,
		"tasks":   taskHeaders,
		"params": gin.H{
			"username":    c.Param("username"),
			"servicename": c.Param("servicename"),
		},
	})
}

func (mc *MainController) ServiceDashboardTab(c *gin.Context) {
	var taskHeaders []*db.GetServiceTasksBySlugRow
	var filteredHeaders []*db.GetServiceTasksBySlugRow

	err := mc.con.store.BeginFunc(c.Request.Context(), func(s *db.Store) error {
		var err error
		taskHeaders, err = s.GetServiceTasksBySlug(
			c.Request.Context(),
			db.GetServiceTasksBySlugParams{
				CreatorUsername: c.Param("username"),
				FormSlug:        c.Param("servicename"),
			},
		)
		if err != nil {
			return err
		}

		// FIXME: create a specialized query to handle this instead
		filteredHeaders = make([]*db.GetServiceTasksBySlugRow, 0)
		for _, t := range taskHeaders {
			if t.Status == db.TaskStatus(c.Param("status")) {
				filteredHeaders = append(filteredHeaders, t)
			}
		}

		return nil
	})
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "serviceDashboardTab.html.tmpl", gin.H{
		"params": gin.H{
			"status":      c.Param("status"),
			"username":    c.Param("username"),
			"servicename": c.Param("servicename"),
		},
		"tasks": filteredHeaders,
	})
}

func (mc *MainController) UpdateTask(c *gin.Context) {
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")
	taskSlug := c.Param("taskname")

	n, err := mc.con.store.UpdateTaskStatus(
		c.Request.Context(),
		db.UpdateTaskStatusParams{
			Status:   db.TaskStatus(c.Query("status")),
			Creator:  serviceCreator,
			FormSlug: serviceSlug,
			TaskSlug: taskSlug,
		},
	)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	} else if len(n) == 0 {
		AbortError(c, http.StatusNotFound, fmt.Errorf("%w: %v", ErrNotFound, serviceCreator))
		return
	}

	c.Status(http.StatusNoContent)
}

func (mc *MainController) ServiceDetail(c *gin.Context) {
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
		mc.con.store,
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

func (mc *MainController) ServiceCreator(c *gin.Context) {
	c.HTML(http.StatusOK, "serviceCreator.html.tmpl", nil)
}

func (mc *MainController) CreateNewService(c *gin.Context) {
	sessionData, ok := GetSessionData(c)
	if !ok {
		AbortError(c, http.StatusUnauthorized, ErrNotLoggedIn)
		return
	}

	var spec NewServiceSpec
	err := c.ShouldBindJSON(&spec)
	if err != nil {
		AbortError(c, http.StatusBadRequest, err)
		return
	}

	f, err := models.CreateServiceForm(
		c.Request.Context(),
		mc.con.store,
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
		fmt.Sprintf("/app/%s/services/%s/dashboard", sessionData.Username, f.Slug))
}

func (mc *MainController) LoginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html.tmpl", nil)
}

func (mc *MainController) RegisterForm(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html.tmpl", nil)
}

func (mc *MainController) HandleLogin(c *gin.Context) {
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
	u, err := mc.con.store.GetUserByUsername(
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
	sessionKey, err := mc.con.sessionManager.NewSession(u.Username, u.ID)
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

func (mc *MainController) HandleRegister(c *gin.Context) {
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

	u, err := mc.con.store.CreateUser(c.Request.Context(), userParams)
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
	sessionKey, err := mc.con.sessionManager.NewSession(u.Username, u.ID)
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
