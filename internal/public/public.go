// Package public provides handler funcs for non-admin API calls to the server
package public

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/bailey4770/chirpy/internal/database"
	"github.com/google/uuid"
)

func HandlerHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK\n")); err != nil {
		log.Fatalf("Error: could not write body to healthz response: %v", err)
	}
}

type chirpParams struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type apiChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type chirpCreator interface {
	CreateChirp(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error)
}

func HandlerPostChirp(db chirpCreator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		chirpReq := chirpParams{}

		if err := json.NewDecoder(req.Body).Decode(&chirpReq); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(chirpReq.Body) > 140 {
			http.Error(w, "Chirp is too long", http.StatusBadRequest)
			return
		}

		chirpDB, err := db.CreateChirp(req.Context(), database.CreateChirpParams{
			Body:   removeProfanity(chirpReq.Body),
			UserID: chirpReq.UserID,
		})
		if err != nil {
			http.Error(w, "Could not create chirp in db", http.StatusInternalServerError)
			return
		}

		chirp := apiChirp{
			ID:        chirpDB.ID,
			CreatedAt: chirpDB.CreatedAt,
			UpdatedAt: chirpDB.UpdatedAt,
			Body:      chirpDB.Body,
			UserID:    chirpDB.UserID,
		}

		log.Printf("Valid chirp from %d received and saved to db", chirp.UserID)
		w.WriteHeader(http.StatusCreated)
		writeResponse(chirp, w)
	}
}

type chirpFetcher interface {
	FetchChirpsByAge(ctx context.Context) ([]database.Chirp, error)
}

func HandlerFetchChirpsByAge(db chirpFetcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		dbChirps, err := db.FetchChirpsByAge(req.Context())
		if err != nil {
			http.Error(w, "could not fetch chirps from db", http.StatusInternalServerError)
			return
		}

		chirps := []apiChirp{}
		for _, dbChirp := range dbChirps {
			chirps = append(chirps, apiChirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.UpdatedAt,
				Body:      dbChirp.Body,
				UserID:    dbChirp.UserID,
			})
		}

		log.Print("All chirps fetched from db")
		w.WriteHeader(http.StatusOK)
		writeResponse(chirps, w)
	}
}

type createUserParams struct {
	Email string `json:"email"`
}

type apiUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type userCreator interface {
	CreateUser(ctx context.Context, email string) (database.User, error)
}

func HandlerCreateUser(db userCreator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		createUserReq := createUserParams{}

		if err := json.NewDecoder(req.Body).Decode(&createUserReq); err != nil {
			log.Printf("Error: could not recode create user request to Go struct: %v", err)
			http.Error(w, "invalid requesy body", http.StatusBadRequest)
			return
		}

		dbUser, err := db.CreateUser(req.Context(), createUserReq.Email)
		if err != nil {
			log.Printf("Error: could not create new user with email %s: %v", createUserReq.Email, err)
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "databse could not create new user with email %s", http.StatusInternalServerError)
			return
		}

		log.Printf("New user %s successfully created", dbUser.Email)

		user := apiUser{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		}

		w.WriteHeader(http.StatusCreated)
		writeResponse(user, w)
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

type responseTypes interface {
	apiChirp | apiUser | []apiChirp
}

func writeResponse[T responseTypes](response T, w http.ResponseWriter) {
	data, err := json.Marshal(&response)
	if err != nil {
		http.Error(w, "Error: could not marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error: could not write response to http body: %v", err)
	}
}
