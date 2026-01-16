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
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	filepathRoot = "./static/"
	port         = "8080"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error: could not load .env file: %v", err)
	}
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error: could not open SQL databse: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		log.Fatalf("Error: could not connect to db: %v", err)
	}

	dbQueries := database.New(db)

	cfg := config.New(dbQueries)
	adminState := &admin.State{DB: dbQueries}
	if os.Getenv("PLATFORM") == "dev" {
		adminState.IsAdmin = true
	}

	mux := http.NewServeMux()
	mux.Handle("/app/",
		adminState.MiddlewareMetricsInc(
			http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))),
		),
	)
	mux.HandleFunc("GET /api/healthz", public.HandlerHealth)
	mux.HandleFunc("POST /api/validate_chirp", public.HandlerValidateChirp)
	mux.HandleFunc("POST /api/users", public.HandlerCreateUser(cfg.DB))

	mux.Handle("GET /admin/metrics", adminState.MiddlewareCheckAdminCreds(adminState.HandlerMetrics))
	mux.Handle("POST /admin/reset", adminState.MiddlewareCheckAdminCreds(adminState.HandlerReset))

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error: could not start listen and serve: %v", err)
	}
}
