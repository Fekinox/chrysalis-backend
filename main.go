package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/config"
	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	config := config.LoadConfig(v, ".env")

	if config.Environment == "dev" {
		pool, res, err := db.InitTestDB(&config)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			fmt.Println("purging database")
			if err := pool.Purge(res); err != nil {
				log.Fatalf("Could not purge database: %s", err)
			}
		}()
	}

	err := db.AutoMigrate(&config)
	if err != nil {
		log.Fatal(err)
	}

	dc, err := CreateController(config)
	if err != nil {
		log.Fatal(err)
	}
	defer dc.Close()

	_, err = NewJSONAPIController(dc)
	if err != nil {
		log.Fatal(err)
	}
	_, err = NewMainController(dc)
	if err != nil {
		log.Fatal(err)
	}

	dc.MountHandlers()

	addr := fmt.Sprintf(":%s", config.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: dc.Router(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a
	// timeout of 5 seconds
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Forced to shutdown: ", err)
	}

	fmt.Println("Server exiting")
}
