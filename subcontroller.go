package main

type JSONAPIController struct {
	c *ChrysalisController
}

func MountJSONAPIController(c *ChrysalisController) (*JSONAPIController, error) {
	con := &JSONAPIController{
		c: c,
	}
	api := c.router.Group("/api")
	api.Use(ErrorHandler(&c.cfg))
	api.Use(SessionKey(c.sessionManager))

	auth := api.Group("/auth")
	_ = auth

	users := api.Group("/users")
	_ = users

	return con, nil
}
