// Package admin provides handler funcs for admin API calls to the server
package admin

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type State struct {
	fileserverHits atomic.Int32
}

const (
	metricsMsg = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
)

func (cfg *State) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *State) HandlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf(metricsMsg, cfg.fileserverHits.Load())
	if _, err := w.Write([]byte(msg)); err != nil {
		log.Fatalf("Error: could not write body to healthz response: %v", err)
	}
}

func (cfg *State) HanlderReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hits reset to 0\n")); err != nil {
		log.Fatalf("Error: could not write to response body: %v", err)
	}
}
