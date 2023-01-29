package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"imgcutter/router"
	"imgcutter/service"

	_ "image/jpeg"
	_ "image/png"
)

func main() {
	services := service.NewService()
	r, err := router.NewRouter(services)
	if err != nil {
		log.Println(err)
		return
	}

	server := http.Server{
		Addr:         ":8080",
		Handler:      r.GetHTTPHandler(),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	log.Printf("starting server...")

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down server")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Println(err)
	}

	if err := services.Session.RemoveAll(); err != nil {
		log.Println(err)
	}
}
