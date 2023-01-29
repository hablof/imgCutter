package main

import (
	"log"
	"net/http"
	"time"

	"imgcutter/router"
	"imgcutter/service"

	_ "image/jpeg"
	_ "image/png"
)

func main() {
	s := service.NewService()
	r, err := router.NewRouter(s)
	if err != nil {
		log.Println(err.Error())
		return
	}

	server := http.Server{
		Addr:         ":8080",
		Handler:      r.GetHTTPHandler(),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	log.Printf("starting server...")

	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
