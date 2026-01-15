package main

import (
	"log"
	"net/http"
)

const (
	filepathRoot = "."
	port         = "8080"
)

func main() {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port %s\n", filepathRoot, port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error: could not start server listen and server: %v", err)
	}
}
