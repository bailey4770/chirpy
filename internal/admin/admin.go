// Package admin provides handler funcs for admin API calls to the server
package admin

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/bailey4770/chirpy/internal/database"
)

type State struct {
	fileserverHits atomic.Int32
	IsAdmin        bool
	DB             *database.Queries
}

const (
	metricsMsg = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
)

func (s *State) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (s *State) MiddlewareCheckAdminCreds(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !s.IsAdmin {
			w.WriteHeader(http.StatusForbidden)
			if _, err := w.Write([]byte("Error: non-admins cannot access admin API\n")); err != nil {
				log.Printf("Error: could not write to response body: %v", err)
			}
		} else {
			f(w, req)
		}
	})
}

func (s *State) HandlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf(metricsMsg, s.fileserverHits.Load())
	if _, err := w.Write([]byte(msg)); err != nil {
		log.Printf("Error: could not write body to healthz response: %v", err)
	}
}

func (s *State) HandlerReset(w http.ResponseWriter, req *http.Request) {
	if err := s.DB.DeleteAllUsers(req.Context()); err != nil {
		log.Printf("Error: could not delete all users from db: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Successfully deleted all users from db\n")); err != nil {
		log.Printf("Error: could not write to response body: %v", err)
	}
}
