package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
	"github.com/Fekinox/chrysalis-backend/internal/genbytes"
	"github.com/Fekinox/chrysalis-backend/internal/models"
	"github.com/Fekinox/chrysalis-backend/internal/session"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type JSONAPIController struct {
	con *ChrysalisServer
}

func (jc *JSONAPIController) MountTo(path string, api gin.IRouter) {
	api.Use(ErrorHandler(&jc.con.cfg, JSONErrorRenderer))
	api.Use(SessionKey(jc.con.sessionManager))
	api.Use(HasCSRFHeader())

	auth := api.Group("/auth")
	auth.POST("/login", jc.Login)
	auth.POST("/register", jc.Register)
	auth.POST("/logout", jc.Logout)
	auth.POST("/csrf", jc.CSRFToken)

	users := api.Group("/users")
	// Get all services a user owns
	users.GET("/:username/services", jc.GetUserServices)
	users.GET("/:username/services/:servicename", jc.GetServiceBySlug)

	users.POST("/:username/services", jc.CreateService)
	users.PUT("/:username/services/:servicename", jc.UpdateService)
	users.DELETE("/:username/services/:servicename", jc.DeleteService)

	// Get outbound and inbound tasks for a user
	users.GET("/:username/outbound-tasks", jc.OutboundTasks)
	users.GET("/:username/inbound-tasks", jc.InboundTasks)

	// Tasks associated with a particular service
	users.GET("/:username/services/:servicename/tasks", jc.GetTasksForService)
	users.GET("/:username/services/:servicename/tasks/:taskslug", jc.GetDetailedTaskInformation)
	users.POST("/:username/services/:servicename/tasks", jc.CreateTaskForService)
	users.POST("/:username/services/:servicename/update", jc.UpdateTasksBulk)
}

func NewJSONAPIController(c *ChrysalisServer) (*JSONAPIController, error) {
	jc := &JSONAPIController{
		con: c,
	}

	c.Mount("/api", jc)

	return jc, nil
}

func (jc *JSONAPIController) Login(c *gin.Context) {
	var loginSchema LoginSchema
	if err := c.ShouldBind(&loginSchema); err != nil {
		AbortError(c, http.StatusBadRequest, ErrBadRequest)
		return
	}

	// Retrieve user from database
	u, err := jc.con.store.GetUserByUsername(
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

	key, ok := GetSessionKey(c)
	if !ok {
		AbortError(c, http.StatusInternalServerError, session.ErrSessionNotFound)
		return
	}

	err = jc.con.sessionManager.Login(key, u.Username, u.ID)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_key": key,
		"username":    u.Username,
		"id":          u.ID,
	})
}

func (jc *JSONAPIController) Register(c *gin.Context) {
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

	u, err := jc.con.store.CreateUser(c.Request.Context(), userParams)
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

	key, ok := GetSessionKey(c)
	if !ok {
		AbortError(c, http.StatusInternalServerError, session.ErrSessionNotFound)
		return
	}

	err = jc.con.sessionManager.Login(key, u.Username, u.ID)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_key": key,
		"username":    u.Username,
		"id":          u.ID,
	})
}

func (jc *JSONAPIController) Logout(c *gin.Context) {
	key, ok := GetSessionKey(c)
	if !ok {
		AbortError(c, http.StatusInternalServerError, session.ErrSessionNotFound)
		return
	}

	err := jc.con.sessionManager.Logout(key)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

func (jc *JSONAPIController) CSRFToken(c *gin.Context) {
	data, ok := GetSessionData(c)
	if !ok {
		AbortError(c, http.StatusForbidden, ErrForbidden)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"csrf-token": hex.EncodeToString(data.CsrfToken),
	})
}

func (jc *JSONAPIController) GetUserServices(c *gin.Context) {
	_ = jc.con.store.BeginFunc(c.Request.Context(), func(s *db.Store) error {
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

func (jc *JSONAPIController) GetServiceBySlug(c *gin.Context) {
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
		jc.con.store,
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

func (jc *JSONAPIController) CreateService(c *gin.Context) {
	username := c.Param("username")

	// Prevent creating a service if the username in the url does not match the
	// logged in user
	if !IsUser(c, username) {
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
		jc.con.store,
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

func (jc *JSONAPIController) UpdateService(c *gin.Context) {
	username := c.Param("username")
	slug := c.Param("servicename")

	if slug == "" {
		AbortError(c, http.StatusBadRequest, errors.New("Missing slug"))
		return
	}

	// Prevent creating a service if the username in the url does not match the
	// logged in user
	if !IsUser(c, username) {
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
		jc.con.store,
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

func (jc *JSONAPIController) DeleteService(c *gin.Context) {
	username := c.Param("username")
	slug := c.Param("servicename")

	if !IsUser(c, username) {
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

	err := jc.con.store.DeleteForm(c.Request.Context(), db.DeleteFormParams{
		Slug:      slug,
		CreatorID: sessionData.UserID,
	})
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func IsUser(c *gin.Context, user string) bool {
	if user == "" {
		return false
	}
	data, ok := GetSessionData(c)
	if !ok {
		return false
	}
	return data.Username == user
}

type CreateTaskParams struct {
	Name    string                      `json:"task_name"`
	Summary string                      `json:"task_summary"`
	Fields  []formfield.FilledFormField `json:"fields"`
}

func generateTaskSlug() (string, error) {
	slug, err := genbytes.GenRandomBytes(4)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", slug), nil
}

// Get all tasks that a user has sent to other services
func (jc *JSONAPIController) OutboundTasks(c *gin.Context) {
	client := c.Param("username")

	task, err := jc.con.store.GetOutboundTasks(c.Request.Context(), client)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Get all tasks that a user has received across all their services
func (dc *JSONAPIController) InboundTasks(c *gin.Context) {
	serviceCreator := c.Param("username")

	task, err := dc.con.store.GetInboundTasks(c.Request.Context(), serviceCreator)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Get all outbound tasks for a specific service
func (jc *JSONAPIController) GetTasksForService(c *gin.Context) {
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")

	task, err := jc.con.store.GetServiceTasksBySlug(
		c.Request.Context(),
		db.GetServiceTasksBySlugParams{
			FormSlug:        serviceSlug,
			CreatorUsername: serviceCreator,
		},
	)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Get detailed information about a task
func (jc *JSONAPIController) GetDetailedTaskInformation(c *gin.Context) {
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")
	taskSlug := c.Param("taskslug")

	task, err := models.GetTask(c.Request.Context(), jc.con.store, models.GetTaskParams{
		CreatorUsername: serviceCreator,
		ServiceName:     serviceSlug,
		TaskName:        taskSlug,
	})
	if err != nil {
		if errors.Is(err,
			errors.Join(
				models.ErrUserNotFound,
				models.ErrServiceNotFound,
				models.ErrFieldsNotFound,
				models.ErrTaskNotFound,
			),
		) {
			AbortError(c, http.StatusNotFound, err)
			return
		}

		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Create an outbound task on a specific service
func (jc *JSONAPIController) CreateTaskForService(c *gin.Context) {
	sessionData, _ := GetSessionData(c)
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")

	if serviceCreator == "" || serviceSlug == "" {
		AbortError(c, http.StatusBadRequest, errors.New("Username or service cannot be empty"))
		return
	}

	var params CreateTaskParams
	err := c.ShouldBindJSON(&params)
	if err != nil {
		AbortError(c, http.StatusBadRequest, err)
		return
	}

	task, err := models.CreateTask(c.Request.Context(), jc.con.store, models.CreateTaskParams{
		CreatorUsername: serviceCreator,
		FormSlug:        serviceSlug,
		ClientID:        sessionData.UserID,
		Fields:          params.Fields,
		TaskName:        params.Name,
		TaskSummary:     params.Summary,
	})
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.ConstraintName == "nonempty_task_name" {
				AbortError(c, http.StatusBadRequest, errors.New("Task name cannot be empty"))
				return
			}
			AbortError(c, http.StatusInternalServerError, err)
			return
		}
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther,
		fmt.Sprintf("/api/users/%s/services/%s/tasks/%s",
			serviceCreator,
			serviceSlug,
			task.TaskSlug),
	)
}

func (jc *JSONAPIController) UpdateTasksBulk(c *gin.Context) {
	var spec []models.TaskUpdateSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		AbortError(c, http.StatusBadRequest, err)
		return
	}

	params := models.UpdateTaskStatusesBulkParams{
		CreatorUsername: c.Param("username"),
		ServiceName:     c.Param("servicename"),
		Updates:         spec,
	}

	var results []models.UpdateTaskStatusesResult
	var err error
	if results, err = models.UpdateTaskStatusesBulk(c.Request.Context(), jc.con.store, params); err != nil {
		if errors.Is(err, models.ErrTaskUpdateDiscrepancies) {
			AbortError(c, http.StatusBadRequest, err)
			return
		}
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, results)
}
