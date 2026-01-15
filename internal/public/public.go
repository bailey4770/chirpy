// Package public provides handler funcs for non-admin API calls to the server
package public

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/bailey4770/chirpy/internal/config"
)

func HandlerHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK\n")); err != nil {
		log.Fatalf("Error: could not write body to healthz response: %v", err)
	}
}

type params struct {
	Body string `json:"body"`
}

type returnVals struct {
	Error       string `json:"error"`
	Valid       bool   `json:"valid"`
	CleanedBody string `json:"cleaned_body"`
}

func HandlerValidateChirp(cfg *config.APIConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		chirp := params{}
		response := returnVals{}

		if err := decoder.Decode(&chirp); err != nil {
			log.Printf("Error: could not decode request to JSON: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			response.Error = "Something went wrong"
		} else if len(chirp.Body) > 140 {
			w.WriteHeader(http.StatusBadRequest)
			response.Error = "Chirp is too long"
		} else {
			w.WriteHeader(http.StatusOK)
			response.Valid = true
			response.CleanedBody = removeProfanity(chirp.Body)
		}

		data, err := json.Marshal(&response)
		if err != nil {
			log.Printf("Error: could not marshal response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(data); err != nil {
			log.Printf("Error: could not write response to http body: %v", err)
		}
	}
}

func removeProfanity(text string) string {
	profanity := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	for _, bw := range profanity {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(bw))
		text = re.ReplaceAllString(text, "****")
	}

	return text
}
