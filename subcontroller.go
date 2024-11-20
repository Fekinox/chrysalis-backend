package main

import "github.com/gin-gonic/gin"

type Subcontroller interface {
	// Given the router interface, mount this subcontroller to that router.
	MountTo(path string, rg gin.IRouter)
}
