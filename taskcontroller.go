package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/genbytes"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func generateTaskSlug() (string, error) {
	slug, err := genbytes.GenRandomBytes(4)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", slug), nil
}

// Get all tasks that a user has sent to other services
func (dc *ChrysalisController) OutboundTasks(c *gin.Context) {
	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	qtx := dc.db.WithTx(tx)

	client := c.Param("username")

	task, err := qtx.GetOutboundTasks(c.Request.Context(), client)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Get all tasks that a user has received across all their services
func (dc *ChrysalisController) InboundTasks(c *gin.Context) {
	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	qtx := dc.db.WithTx(tx)

	serviceCreator := c.Param("username")

	task, err := qtx.GetInboundTasks(c.Request.Context(), serviceCreator)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Get all outbound tasks for a specific service
func (dc *ChrysalisController) GetTasksForService(c *gin.Context) {
	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	qtx := dc.db.WithTx(tx)

	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")

	task, err := qtx.GetServiceTasksBySlug(
		c.Request.Context(),
		db.GetServiceTasksBySlugParams{
			FormSlug:        serviceSlug,
			CreatorUsername: serviceCreator,
		},
	)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Get detailed information about a task
func (dc *ChrysalisController) GetDetailedTaskInformation(c *gin.Context) {
	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	qtx := dc.db.WithTx(tx)

	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")
	taskSlug := c.Param("taskslug")

	task, err := qtx.GetTaskHeader(c.Request.Context(), db.GetTaskHeaderParams{
		Username: serviceCreator,
		FormSlug: serviceSlug,
		TaskSlug: taskSlug,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Create an outbound task on a specific service
func (dc *ChrysalisController) CreateTaskForService(c *gin.Context) {
	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	qtx := dc.db.WithTx(tx)

	sessionData, _ := GetSessionData(c)
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")

	if serviceCreator == "" || serviceSlug == "" {
		c.AbortWithError(
			http.StatusNotFound,
			errors.New("Username or service cannot be empty"),
		)
		return
	}

	creator, err := qtx.GetUserByUsername(c.Request.Context(), serviceCreator)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	form, err := qtx.GetCurrentFormVersionBySlug(
		c.Request.Context(),
		db.GetCurrentFormVersionBySlugParams{
			Slug:      serviceSlug,
			CreatorID: creator.ID,
		},
	)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var task *db.CreateTaskRow
	for {
		taskSlug, err := generateTaskSlug()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		task, err = qtx.CreateTask(c.Request.Context(), db.CreateTaskParams{
			FormVersionID: form.FormVersionID,
			ClientID:      sessionData.UserID,
			Slug:          taskSlug,
		})
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				continue
			} else {
				c.AbortWithError(http.StatusInternalServerError, pgErr)
			}
			return
		} else if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		break
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, task)
}

// Update the status of a task as the owner of a service
func (dc *ChrysalisController) UpdateTask(c *gin.Context) {
	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c.Request.Context())
	qtx := dc.db.WithTx(tx)

	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")
	taskSlug := c.Param("taskslug")

	n, err := qtx.UpdateTaskStatus(c.Request.Context(), db.UpdateTaskStatusParams{
		Status: db.TaskStatus(c.Query("status")),
		Creator: serviceCreator,
		FormSlug: serviceSlug,
		TaskSlug: taskSlug,
	})
	fmt.Println(n)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if len(n) == 0 {
		c.AbortWithError(http.StatusNotFound, ErrNotFound(serviceCreator))
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
