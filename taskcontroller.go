package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
	"github.com/Fekinox/chrysalis-backend/internal/genbytes"
	"github.com/Fekinox/chrysalis-backend/internal/models"
	"github.com/gin-gonic/gin"
)

type CreateTaskParams struct {
	Fields []formfield.FilledFormField `json:"fields"`
}

func generateTaskSlug() (string, error) {
	slug, err := genbytes.GenRandomBytes(4)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", slug), nil
}

// Get all tasks that a user has sent to other services
func (dc *ChrysalisController) OutboundTasks(c *gin.Context) {
	client := c.Param("username")

	task, err := dc.store.GetOutboundTasks(c.Request.Context(), client)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Get all tasks that a user has received across all their services
func (dc *ChrysalisController) InboundTasks(c *gin.Context) {
	serviceCreator := c.Param("username")

	task, err := dc.store.GetInboundTasks(c.Request.Context(), serviceCreator)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Get all outbound tasks for a specific service
func (dc *ChrysalisController) GetTasksForService(c *gin.Context) {
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")

	task, err := dc.store.GetServiceTasksBySlug(
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
func (dc *ChrysalisController) GetDetailedTaskInformation(c *gin.Context) {
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")
	taskSlug := c.Param("taskslug")

	task, err := models.GetTask(c.Request.Context(), dc.store, models.GetTaskParams{
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
func (dc *ChrysalisController) CreateTaskForService(c *gin.Context) {
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

	task, err := models.CreateTask(c.Request.Context(), dc.store, models.CreateTaskParams{
		CreatorUsername: serviceCreator,
		FormSlug:        serviceSlug,
		ClientID:        sessionData.UserID,
		Fields:          params.Fields,
	})
	if err != nil {
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

// Update the status of a task as the owner of a service
func (dc *ChrysalisController) UpdateTask(c *gin.Context) {
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")
	taskSlug := c.Param("taskslug")

	n, err := dc.store.UpdateTaskStatus(
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

	c.Redirect(http.StatusSeeOther,
		fmt.Sprintf(
			"/api/users/%s/services/%s/tasks/%s",
			serviceCreator,
			serviceSlug,
			taskSlug,
		),
	)
}
