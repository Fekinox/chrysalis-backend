package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
	"github.com/Fekinox/chrysalis-backend/internal/genbytes"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

const MAX_TASK_RETRY_ATTEMPTS int = 10

func generateTaskSlug() (string, error) {
	slug, err := genbytes.GenRandomBytes(4)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", slug), nil
}

// Get all tasks that a user has sent to other services
func (dc *ChrysalisController) OutboundTasks(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		client := c.Param("username")

		task, err := s.GetOutboundTasks(c.Request.Context(), client)
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, task)
		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

// Get all tasks that a user has received across all their services
func (dc *ChrysalisController) InboundTasks(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		serviceCreator := c.Param("username")

		task, err := s.GetInboundTasks(c.Request.Context(), serviceCreator)
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, task)
		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

// Get all outbound tasks for a specific service
func (dc *ChrysalisController) GetTasksForService(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		serviceCreator := c.Param("username")
		serviceSlug := c.Param("servicename")

		task, err := s.GetServiceTasksBySlug(
			c.Request.Context(),
			db.GetServiceTasksBySlugParams{
				FormSlug:        serviceSlug,
				CreatorUsername: serviceCreator,
			},
		)
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, task)
		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

// Get detailed information about a task
func (dc *ChrysalisController) GetDetailedTaskInformation(c *gin.Context) {
	err := dc.StoreFuncTx(c.Request.Context(), func(s *db.Store) error {
		serviceCreator := c.Param("username")
		serviceSlug := c.Param("servicename")
		taskSlug := c.Param("taskslug")

		task, err := s.GetTaskHeader(
			c.Request.Context(),
			db.GetTaskHeaderParams{
				Username: serviceCreator,
				FormSlug: serviceSlug,
				TaskSlug: taskSlug,
			},
		)
		if err != nil {
			return err
		}

		rawFields, err := s.GetFilledFormFields(c.Request.Context(), taskSlug)
		if err != nil {
			return err
		}

		parsedFields := make([]formfield.FilledFormField, len(rawFields))

		for i, f := range rawFields {
			err = parsedFields[i].FromRow(f)
			if err != nil {
				return err
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"taskInfo": task,
			"fields":   parsedFields,
		})
		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

// Create an outbound task on a specific service
func (dc *ChrysalisController) CreateTaskForService(c *gin.Context) {
	err := dc.StoreFuncTx(c.Request.Context(), func(s *db.Store) error {
		sessionData, _ := GetSessionData(c)
		serviceCreator := c.Param("username")
		serviceSlug := c.Param("servicename")

		if serviceCreator == "" || serviceSlug == "" {
			return errors.New("Username or service cannot be empty")
		}

		creator, err := s.GetUserByUsername(c.Request.Context(), serviceCreator)
		if err != nil {
			return err
		}

		form, err := s.GetCurrentFormVersionBySlug(
			c.Request.Context(),
			db.GetCurrentFormVersionBySlugParams{
				Slug:      serviceSlug,
				CreatorID: creator.ID,
			},
		)
		if err != nil {
			return err
		}

		var task *db.CreateTaskRow
		var attempts int
		for {
			err := s.BeginFunc(
				c.Request.Context(),
				func(loopTx *db.Store) error {
					taskSlug, err := generateTaskSlug()
					if err != nil {
						return err
					}

					task, err = loopTx.
						CreateTask(c.Request.Context(), db.CreateTaskParams{
							FormVersionID: form.FormVersionID,
							ClientID:      sessionData.UserID,
							Slug:          taskSlug,
						})

					return err
				},
			)

			var pgErr *pgconn.PgError

			if err == nil {
				break
			} else if errors.As(err, &pgErr) {
				if pgErr.Code != "23505" || pgErr.ConstraintName != "task_slug_unique" {
					return err
				}
			} else {
				return err
			}

			attempts++

			if attempts >= MAX_TASK_RETRY_ATTEMPTS {
				return fmt.Errorf("Too many retry attempts")
			}

			time.Sleep(time.Millisecond * 50)
		}

		c.JSON(http.StatusCreated, task)

		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}

// Update the status of a task as the owner of a service
func (dc *ChrysalisController) UpdateTask(c *gin.Context) {
	err := dc.StoreFunc(c.Request.Context(), func(s *db.Store) error {
		serviceCreator := c.Param("username")
		serviceSlug := c.Param("servicename")
		taskSlug := c.Param("taskslug")

		n, err := s.UpdateTaskStatus(
			c.Request.Context(),
			db.UpdateTaskStatusParams{
				Status:   db.TaskStatus(c.Query("status")),
				Creator:  serviceCreator,
				FormSlug: serviceSlug,
				TaskSlug: taskSlug,
			},
		)
		if err != nil {
			return err
		} else if len(n) == 0 {
			return ErrNotFound(serviceCreator)
		}

		c.Redirect(http.StatusSeeOther,
			fmt.Sprintf(
				"/api/users/%s/services/%s/tasks/%s",
				serviceCreator,
				serviceSlug,
				taskSlug,
			),
		)
		return nil
	})

	// TODO: Implement proper error codes
	if err != nil {
		code := http.StatusInternalServerError
		AbortError(c, code, err)
	}
}
