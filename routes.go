package main

import "github.com/gin-contrib/cors"

func (dc *ChrysalisServer) Mount(path string, sc Subcontroller) {
	g := dc.router.Group(path)
	sc.MountTo(path, g)
}

func (dc *ChrysalisServer) MountHandlers() {
	dc.router.Use(cors.New(dc.cors))
	dc.router.GET("/", dc.Healthcheck)
	dc.router.GET("/healthcheck-inner", HTMXRedirect("/"), dc.HealthcheckInner)
	dc.router.GET("/healthcheck-objects", HTMXRedirect("/"), dc.HealthcheckObjectStats)
}
