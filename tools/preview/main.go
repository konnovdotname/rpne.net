package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Сайт: http://127.0.0.1:8081/")
	log.Fatal(http.ListenAndServe(
		"127.0.0.1:8081",
		http.FileServer(http.Dir(root)),
	))
}
