package main

import (
	"database/sql"
	"fmt"
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
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer func() { _ = db.Close() }()

	cfg, adminState := loadConfigs(db)

	mux := http.NewServeMux()
	registerRoutes(mux, cfg, adminState)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error: could not start listen and serve: %v", err)
	}
}

func connectDB() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("could not load .env file: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("could not open SQL databse: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to db: %v", err)
	}

	return db, nil
}

func loadConfigs(db *sql.DB) (*config.APIConfig, *admin.State) {
	dbQueries := database.New(db)

	cfg := &config.APIConfig{DB: dbQueries}
	cfg.Secret = os.Getenv("SECRET")

	adminState := &admin.State{DB: dbQueries}
	if os.Getenv("PLATFORM") == "dev" {
		adminState.IsAdmin = true
	}

	return cfg, adminState
}

func registerRoutes(mux *http.ServeMux, cfg *config.APIConfig, adminState *admin.State) {
	mux.Handle("/app/",
		adminState.MiddlewareMetricsInc(
			http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))),
		),
	)
	mux.HandleFunc("GET /api/healthz", public.HandlerHealth)
	mux.HandleFunc("GET /api/chirps", public.HandlerFetchChirpsByAge(cfg.DB))
	mux.HandleFunc("GET /api/chirps/{chirpID}", public.HandlerFetchChirpByID(cfg.DB))
	mux.HandleFunc("POST /api/chirps", public.HandlerPostChirp(cfg.DB, cfg.Secret))
	mux.HandleFunc("POST /api/users", public.HandlerCreateUser(cfg.DB))
	mux.HandleFunc("POST /api/login", public.HandlerLogin(cfg.DB, cfg.Secret))

	mux.Handle("GET /admin/metrics", adminState.MiddlewareCheckAdminCreds(adminState.HandlerMetrics))
	mux.Handle("POST /admin/reset", adminState.MiddlewareCheckAdminCreds(adminState.HandlerReset))
}
