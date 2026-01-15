// Package apiconfig provides handler funcs for API calls to the server
package apiconfig

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type APIConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *APIConfig) HandlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	if _, err := w.Write([]byte(msg)); err != nil {
		log.Fatalf("Error: could not write body to healthz response: %v", err)
	}
}

func (cfg *APIConfig) HanlderReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hits reset to 0")); err != nil {
		log.Fatalf("Error: could not write to response body: %v", err)
	}
}
