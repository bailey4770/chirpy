package main

import (
	"log"
	"net/http"

	"github.com/bailey4770/chirpy/internal/apiconfig"
)

const (
	filepathRoot = "./static/"
	port         = "8080"
)

func main() {
	cfg := apiconfig.APIConfig{}

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/",
		cfg.MiddlewareMetricsInc(
			http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))),
		),
	)
	serveMux.HandleFunc("/healthz", handlerHealth)
	serveMux.HandleFunc("/metrics", cfg.HandlerMetrics)
	serveMux.HandleFunc("/reset", cfg.HanlderReset)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error: could not start server listen and server: %v", err)
	}
}

func handlerHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Fatalf("Error: could not write body to healthz response: %v", err)
	}
}
