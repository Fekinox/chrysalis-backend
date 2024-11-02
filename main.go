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

	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	config := LoadConfig(v, ".env")

	dc, err := CreateController(config)
	if err != nil {
		log.Fatal(err)
	}
	defer dc.Close()

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
