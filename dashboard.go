package main

import (
	"fmt"
	"net/http"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/models"
	"github.com/gin-gonic/gin"
)

func (dc *ChrysalisController) UserDashboard(c *gin.Context) {
	sessionData, _ := GetSessionData(c)

	services, err := dc.store.GetUserFormHeaders(c.Request.Context(), sessionData.UserID)
	if err != nil {
		AbortError(c, http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "userDashboard.html.tmpl", gin.H{
		"session":  sessionData,
		"services": services,
	})
}

func (dc *ChrysalisController) ServiceDashboard(c *gin.Context) {
	sessionData, _ := GetSessionData(c)

	var form *models.ServiceForm
	var taskHeaders []*db.GetServiceTasksBySlugRow

	err := dc.store.BeginFunc(c.Request.Context(), func(s *db.Store) error {
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

func (dc *ChrysalisController) ServiceDashboardTab(c *gin.Context) {
	var taskHeaders []*db.GetServiceTasksBySlugRow
	var filteredHeaders []*db.GetServiceTasksBySlugRow

	err := dc.store.BeginFunc(c.Request.Context(), func(s *db.Store) error {
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

func (dc *ChrysalisController) UpdateTaskDashboard(c *gin.Context) {
	serviceCreator := c.Param("username")
	serviceSlug := c.Param("servicename")
	taskSlug := c.Param("taskname")

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

	c.Status(http.StatusNoContent)
}
