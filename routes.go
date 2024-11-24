package main

func (dc *ChrysalisServer) Mount(path string, sc Subcontroller) {
	g := dc.router.Group(path)
	sc.MountTo(path, g)
}

func (dc *ChrysalisServer) MountHandlers() {
	dc.router.GET("/", dc.Healthcheck)
	dc.router.GET("/healthcheck-inner", HTMXRedirect("/"), dc.HealthcheckInner)
	dc.router.GET("/healthcheck-objects", HTMXRedirect("/"), dc.HealthcheckObjectStats)
}
