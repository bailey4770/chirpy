package main

import (
	"log"
	"net/http"

	"github.com/bailey4770/chirpy/internal/adminapi"
	"github.com/bailey4770/chirpy/internal/api"
)

const (
	filepathRoot = "./static/"
	port         = "8080"
)

func main() {
	cfg := adminapi.APIConfig{}

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/",
		cfg.MiddlewareMetricsInc(
			http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))),
		),
	)
	serveMux.HandleFunc("GET /api/healthz", api.HandlerHealth)
	serveMux.HandleFunc("POST /api/validate_chirp", api.HandlerValidateChirp)
	serveMux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	serveMux.HandleFunc("POST /admin/reset", cfg.HanlderReset)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error: could not start server listen and server: %v", err)
	}
}
