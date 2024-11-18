package main

import (
	"net/http"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/models"
	"github.com/gin-gonic/gin"
)

func (dc *ChrysalisController) UserDashboard(c *gin.Context) {
	sessionData, _ := GetSessionData(c)

	c.HTML(http.StatusOK, "userDashboard.html.tmpl", gin.H{
		"session": sessionData,
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
			Service: c.Param("servicename"),
		})
		if err != nil {
			return err
		}

		taskHeaders, err = s.GetServiceTasksBySlug(c.Request.Context(), db.GetServiceTasksBySlugParams{
			CreatorUsername: c.Param("username"),
			FormSlug: c.Param("servicename"),
		})
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
		"tasks": taskHeaders,
	})
}
