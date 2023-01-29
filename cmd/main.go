package main

import (
	"log"
	"net/http"

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

	log.Printf("starting server...")
	_ = http.ListenAndServe(":8080", r.GetHTTPHandler())
}
