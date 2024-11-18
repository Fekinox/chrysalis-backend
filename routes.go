package main

func (dc *ChrysalisController) MountHandlers() {
	api := dc.router.Group("/api")
	api.Use(ErrorHandler(&dc.cfg))
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

	// HTML api
	app := dc.router.Group("/app")
	app.Use(ErrorHandler(&dc.cfg))
	app.Use(SessionKey(dc.sessionManager))
	app.GET("/helloworld", dc.DummyTemplateHandler)
	app.GET("/:username/services", dc.GetUserServicesHTML)
	app.GET("/:username/services/:servicename/dashboard", dc.ServiceDashboard)
	app.GET(
		"/:username/services/:servicename/form",
		RedirectToLogin(dc.sessionManager),
		dc.GetServiceDetail,
	)
	app.GET("/:username/services/:servicename/edit",
		RedirectToLogin(dc.sessionManager),
		dc.GetServiceDetail,
	)

	app.GET("/new-service",
		RedirectToLogin(dc.sessionManager),
		dc.ServiceCreator,
	)
	app.POST("/new-service",
		HasSessionKey(dc.sessionManager),
		dc.CreateNewService,
	)
	app.GET("/:username/services/:servicename/tasks/:taskname", dc.DummyTemplateHandler)

	app.GET("/login", dc.LoginForm)
	app.POST("/login", dc.HandleLogin)

	app.GET("/register", dc.RegisterForm)
	app.POST("/register", dc.HandleRegister)

	app.GET("/dashboard", RedirectToLogin(dc.sessionManager), dc.UserDashboard)
}
