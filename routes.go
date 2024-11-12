package main

func (dc *ChrysalisController) MountHandlers() {
	dc.router.Use(ErrorHandler(&dc.cfg))

	api := dc.router.Group("/api")
	api.Use(SessionKey(dc.sessionManager))

	auth := api.Group("/auth")
	auth.POST("/login", dc.Login)
	auth.POST("/register", dc.Register)
	auth.POST("/logout", HasSessionKey(dc.sessionManager), dc.Logout)

	users := api.Group("/users")
	// Get all services a user owns
	users.GET("/:username/services", dc.GetUserServices)
	users.GET("/:username/services/:servicename", dc.GetServiceBySlug)

	users.POST("/:username/services",
		HasSessionKey(dc.sessionManager),
		dc.CreateService)
	users.PUT("/:username/services/:servicename",
		HasSessionKey(dc.sessionManager),
		dc.UpdateService)
	users.DELETE("/:username/services/:servicename",
		HasSessionKey(dc.sessionManager),
		dc.DeleteService)

	// Get outbound and inbound tasks for a user
	users.GET("/:username/outbound-tasks", dc.OutboundTasks)
	users.GET("/:username/inbound-tasks", dc.InboundTasks)

	// Tasks associated with a particular service
	users.GET("/:username/services/:servicename/tasks", dc.GetTasksForService)
	users.GET(
		"/:username/services/:servicename/tasks/:taskslug",
		dc.GetDetailedTaskInformation,
	)
	users.POST(
		"/:username/services/:servicename/tasks",
		HasSessionKey(dc.sessionManager),
		dc.CreateTaskForService,
	)
	users.PUT(
		"/:username/services/:servicename/tasks/:taskslug",
		dc.UpdateTask,
	)

	dc.router.GET("/helloworld", dc.DummyTemplateHandler)
	dc.router.GET("/:username/services", dc.GetUserServicesHTML)
	dc.router.GET("/:username/services/:servicename", dc.GetServiceDetail)
}
