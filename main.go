package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/bailey4770/chirpy/internal/admin"
	"github.com/bailey4770/chirpy/internal/config"
	"github.com/bailey4770/chirpy/internal/database"
	"github.com/bailey4770/chirpy/internal/public"
	_ "github.com/lib/pq"
)

const (
	filepathRoot = "./static/"
	port         = "8080"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error: could not open SQL databse: %v", err)
	}

	dbQueries := database.New(db)
	cfg := config.New(dbQueries)
	adminState := &admin.State{}

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/",
		adminState.MiddlewareMetricsInc(
			http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))),
		),
	)
	serveMux.HandleFunc("GET /api/healthz", public.HandlerHealth)
	serveMux.HandleFunc("POST /api/validate_chirp", public.HandlerValidateChirp(cfg))
	serveMux.HandleFunc("GET /admin/metrics", adminState.HandlerMetrics)
	serveMux.HandleFunc("POST /admin/reset", adminState.HanlderReset)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error: could not start server listen and server: %v", err)
	}
}
