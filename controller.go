package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	db "github.com/Fekinox/chrysalis-backend/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type ChrysalisController struct {
	db     *db.Queries
	cfg    Config
	router *gin.Engine
	conn   *pgx.Conn
}

var (
	BadRequestError = errors.New("Bad request")
	NotFoundError   = func(name string) error {
		return errors.New(fmt.Sprintf("Not found: %s", name))
	}
)

func CreateController(cfg Config) (*ChrysalisController, error) {
	var engine *gin.Engine
	if cfg.Environment == "test" || cfg.Environment == "release" {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
	} else {
		engine = gin.Default()
	}

	conn, err := pgx.Connect(context.Background(), cfg.GetDBUrl())
	if err != nil {
		return nil, err
	}

	q := db.New(conn)

	return &ChrysalisController{
		db:     q,
		cfg:    cfg,
		conn:   conn,
		router: engine,
	}, nil
}

func (dc *ChrysalisController) MountHandlers() {
	api := dc.router.Group("/api")
	api.Use(ErrorHandler(&dc.cfg))

	auth := api.Group("/auth")
	auth.POST("/login", dc.DummyHandler)
	auth.POST("/register", dc.DummyHandler)

	users := api.Group("/users")
	// Get all services a user owns
	users.GET("/:username/services", dc.DummyHandler)
	users.GET("/:username/services/:servicename", dc.DummyHandler)
	users.POST("/:username/services", dc.DummyHandler)
	users.PUT("/:username/services/:servicename", dc.DummyHandler)
	users.DELETE("/:username/services/:servicename", dc.DummyHandler)

	// Get outbound and inbound tasks for a user
	users.GET("/:username/outbound-tasks", dc.DummyHandler)
	users.GET("/:username/inbound-tasks", dc.DummyHandler)

	// Tasks associated with a particular service
	users.GET("/:username/services/:servicename/tasks", dc.DummyHandler)
	users.GET("/:username/services/:servicename/tasks/:taskslug", dc.DummyHandler)
	users.POST("/:username/services/:servicename/tasks", dc.DummyHandler)
	users.PUT("/:username/services/:servicename/tasks/:taskslug", dc.DummyHandler)
}

func (dc *ChrysalisController) Start(addr string) error {
	return dc.router.Run(addr)
}

func (dc *ChrysalisController) Router() *gin.Engine {
	return dc.router
}

func (dc *ChrysalisController) Close() error {
	return dc.conn.Close(context.Background())
}

func (dc *ChrysalisController) DummyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"method": c.Request.Method,
		"url": c.Request.URL.RequestURI(),
	})
}
