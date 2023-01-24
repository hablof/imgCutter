package main

import (
	"imgCutter/router"
	"imgCutter/service"
	"log"
	"net/http"

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

	log.Printf("starting server...")
	http.ListenAndServe(":8080", r.GetHTTPHandler())
}
