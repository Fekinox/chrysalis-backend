package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (dc *ChrysalisController) UserDashboard(c *gin.Context) {
	sessionData, ok := GetSessionData(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/app/login")
	}

	c.HTML(http.StatusOK, "serviceDashboard.html.tmpl", gin.H{
		"session": sessionData,
	})
}
