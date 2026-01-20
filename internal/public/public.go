// Package public provides handler funcs for non-admin API calls to the server
package public

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/bailey4770/chirpy/internal/auth"
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
	Body string `json:"body"`
}

type apiChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func dbChirpToAPIChirp(dbChirp database.Chirp) apiChirp {
	return apiChirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}

type chirpCreator interface {
	CreateChirp(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error)
}

func HandlerPostChirp(db chirpCreator, secret string) func(http.ResponseWriter, *http.Request) {
	removeProfanity := func(text string) string {
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

	return func(w http.ResponseWriter, req *http.Request) {
		chirpReq := chirpParams{}

		if err := json.NewDecoder(req.Body).Decode(&chirpReq); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		tokenUserID, err := auth.ValidateJWT(token, secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if len(chirpReq.Body) > 140 {
			http.Error(w, "Chirp is too long", http.StatusBadRequest)
			return
		}

		dbChirp, err := db.CreateChirp(req.Context(), database.CreateChirpParams{
			Body:   removeProfanity(chirpReq.Body),
			UserID: tokenUserID,
		})
		if err != nil {
			http.Error(w, "Could not create chirp in db", http.StatusInternalServerError)
			return
		}

		chirp := dbChirpToAPIChirp(dbChirp)

		log.Printf("Valid chirp from %d received and saved to db", chirp.UserID)
		w.WriteHeader(http.StatusCreated)
		writeResponse(chirp, w)
	}
}

type allChirpsFetcher interface {
	FetchChirpsByAge(ctx context.Context) ([]database.Chirp, error)
}

func HandlerFetchChirpsByAge(db allChirpsFetcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		dbChirps, err := db.FetchChirpsByAge(req.Context())
		if err != nil {
			http.Error(w, "could not fetch chirps from db", http.StatusInternalServerError)
			return
		}

		chirps := []apiChirp{}
		for _, dbChirp := range dbChirps {
			chirps = append(chirps, dbChirpToAPIChirp(dbChirp))
		}

		log.Print("All chirps fetched from db")
		w.WriteHeader(http.StatusOK)
		writeResponse(chirps, w)
	}
}

type chirpFetcher interface {
	FetchChirpByID(ctx context.Context, id uuid.UUID) (database.Chirp, error)
}

func HandlerFetchChirpByID(db chirpFetcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		chirpID, err := uuid.Parse(req.PathValue("chirpID"))
		if err != nil {
			http.Error(w, "could not parse chirp ID to uuid", http.StatusBadRequest)
			return
		}

		dbChirp, err := db.FetchChirpByID(req.Context(), chirpID)
		if err != nil {
			http.Error(w, "could not fetch requested chirp", http.StatusNotFound)
			return
		}

		chirp := dbChirpToAPIChirp(dbChirp)

		log.Print("Chirp fetched from db")
		w.WriteHeader(http.StatusOK)
		writeResponse(chirp, w)
	}
}

type userRequestParams struct {
	Password  string        `json:"password"`
	Email     string        `json:"email"`
	ExpiresIn time.Duration `json:"expires_in_seconds"`
}

type apiUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

type userCreator interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.CreateUserRow, error)
}

func HandlerCreateUser(db userCreator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		createUserReq := userRequestParams{}

		if err := json.NewDecoder(req.Body).Decode(&createUserReq); err != nil {
			log.Printf("Error: could not recode create user request to Go struct: %v", err)
			http.Error(w, "invalid requesy body", http.StatusBadRequest)
			return
		}

		hashedPassword, err := auth.HashPassword(createUserReq.Password)
		if err != nil {
			http.Error(w, "could not hash provided password", http.StatusInternalServerError)
			return
		}

		dbUser, err := db.CreateUser(req.Context(), database.CreateUserParams{
			Email:          createUserReq.Email,
			HashedPassword: hashedPassword,
		})
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

type userRetreiver interface {
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
}

func HandlerLogin(db userRetreiver, secret string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		loginReq := userRequestParams{}

		if err := json.NewDecoder(req.Body).Decode(&loginReq); err != nil {
			log.Printf("Error: could not decode json: %v", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		dbUser, err := db.GetUserByEmail(req.Context(), loginReq.Email)
		if err != nil {
			http.Error(w, "Incorrect email or password", http.StatusUnauthorized)
			return
		}

		ok, err := auth.CheckPasswordHash(loginReq.Password, dbUser.HashedPassword)
		if err != nil || !ok {
			http.Error(w, "Incorrect email or password", http.StatusUnauthorized)
			return
		}

		log.Printf("User %s successfully logged in", dbUser.Email)

		expires := loginReq.ExpiresIn
		if loginReq.ExpiresIn > time.Hour || loginReq.ExpiresIn == time.Duration(0) {
			expires = time.Duration(time.Hour)
		}

		token, err := auth.MakeJWT(dbUser.ID, secret, expires)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		user := apiUser{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
			Token:     token,
		}

		w.WriteHeader(http.StatusOK)
		writeResponse(user, w)
	}
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
