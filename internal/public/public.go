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
			http.Error(w, "could not get bearer token from header", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateJWT(token, secret)
		if err != nil {
			log.Printf("Error: could not validate JWT: %v", err)
			http.Error(w, "could not validate JWT", http.StatusUnauthorized)
			return
		}

		if len(chirpReq.Body) > 140 {
			http.Error(w, "Chirp is too long", http.StatusBadRequest)
			return
		}

		dbChirp, err := db.CreateChirp(req.Context(), database.CreateChirpParams{
			Body:   removeProfanity(chirpReq.Body),
			UserID: userID,
		})
		if err != nil {
			http.Error(w, "Could not create chirp in db", http.StatusInternalServerError)
			return
		}

		chirp := dbChirpToAPIChirp(dbChirp)

		log.Printf("User %v successfully posted chirp %v", userID, dbChirp.ID)
		w.WriteHeader(http.StatusCreated)
		writeResponse(chirp, w)
	}
}

type chirpStore interface {
	FetchChirpsByAge(ctx context.Context) ([]database.Chirp, error)
	FetchChirpByID(ctx context.Context, id uuid.UUID) (database.Chirp, error)
	DeleteChirp(ctx context.Context, id uuid.UUID) error
}

func HandlerFetchChirpsByAge(db chirpStore) func(http.ResponseWriter, *http.Request) {
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

		w.WriteHeader(http.StatusOK)
		writeResponse(chirps, w)
	}
}

func HandlerFetchChirpByID(db chirpStore) func(http.ResponseWriter, *http.Request) {
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

		log.Printf("Chirp %v successfully requested ", chirpID)
		w.WriteHeader(http.StatusOK)
		writeResponse(chirp, w)
	}
}

func HandlerDeleteChirp(db chirpStore, secret string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(w, "could not get bearer token from header", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateJWT(token, secret)
		if err != nil {
			log.Printf("Error: could not validate JWT: %v", err)
			http.Error(w, "could not validate JWT", http.StatusUnauthorized)
			return
		}

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

		if userID != dbChirp.UserID {
			http.Error(w, "request user ID does not match chirp's user ID", http.StatusForbidden)
			return
		}

		if err = db.DeleteChirp(req.Context(), dbChirp.ID); err != nil {
			log.Printf("Error: could not delete chirp from db: %v", err)
			http.Error(w, "could no delete chirp from db", http.StatusInternalServerError)
			return
		}

		log.Printf("Warning: chirp %v successfully deleted by %v", chirpID, userID)
		w.WriteHeader(http.StatusNoContent)
	}
}

type userRequestParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type apiUser struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
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
			http.Error(w, "databse could not create new user with email %s", http.StatusInternalServerError)
			return
		}

		log.Printf("New user %s successfully created", dbUser.Email)

		user := apiUser{
			ID:          dbUser.ID,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
			Email:       dbUser.Email,
			IsChirpyRed: dbUser.IsChirpyRed,
		}

		w.WriteHeader(http.StatusCreated)
		writeResponse(user, w)
	}
}

type polkaWebhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

type userUpgrader interface {
	MakeUserRed(ctx context.Context, id uuid.UUID) error
}

func HandlerUpgradeUser(db userUpgrader, polkaKey string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		apiKey, err := auth.GetAPIKey(req.Header)
		if err != nil || apiKey != polkaKey {
			log.Printf("Error: %v", err)
			http.Error(w, "invalid api key in authorisation header", http.StatusUnauthorized)
			return
		}

		upgradeReq := polkaWebhook{}
		if err := json.NewDecoder(req.Body).Decode(&upgradeReq); err != nil {
			log.Printf("Error: could not decode json: %v", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if upgradeReq.Event != "user.upgraded" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := db.MakeUserRed(req.Context(), upgradeReq.Data.UserID); err != nil {
			log.Printf("Error: could not upgrade user %v: %v", upgradeReq.Data.UserID, err)
			http.Error(w, "could not find user", http.StatusNotFound)
			return
		}

		log.Printf("User %v succesffully upgraded to red", upgradeReq.Data.UserID)
		w.WriteHeader(http.StatusNoContent)
	}
}

type authStore interface {
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	UpdateEmailAndPassword(ctx context.Context, arg database.UpdateEmailAndPasswordParams) (database.UpdateEmailAndPasswordRow, error)
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.RefreshToken, error)
	GetRefreshToken(ctx context.Context, token string) (database.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

func HandlerLogin(db authStore, secret string) func(http.ResponseWriter, *http.Request) {
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

		token, err := auth.MakeJWT(dbUser.ID, secret)
		if err != nil {
			log.Printf("Error: could not make JWT: %v", err)
			http.Error(w, "could not make JWT", http.StatusInternalServerError)
			return
		}

		refreshToken := auth.MakeRefreshToken()
		_, err = db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    dbUser.ID,
			ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		})
		if err != nil {
			log.Printf("Error: could not create refresh token: %v", err)
			http.Error(w, "could not create refresh token", http.StatusInternalServerError)
			return
		}

		user := apiUser{
			ID:           dbUser.ID,
			CreatedAt:    dbUser.CreatedAt,
			UpdatedAt:    dbUser.UpdatedAt,
			Email:        dbUser.Email,
			Token:        token,
			RefreshToken: refreshToken,
			IsChirpyRed:  dbUser.IsChirpyRed,
		}

		w.WriteHeader(http.StatusOK)
		writeResponse(user, w)
	}
}

func HandlerUpdateEmailAndPassword(db authStore, secret string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(w, "could not get bearer token from header", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateJWT(token, secret)
		if err != nil {
			log.Printf("Error: could not validate JWT: %v", err)
			http.Error(w, "could not validate JWT", http.StatusUnauthorized)
			return
		}

		var userReq userRequestParams
		if err := json.NewDecoder(req.Body).Decode(&userReq); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		hashedPassword, err := auth.HashPassword(userReq.Password)
		if err != nil {
			log.Printf("Error: could not hash password: %v", err)
			http.Error(w, "could not hash password", http.StatusInternalServerError)
			return
		}

		dbUser, err := db.UpdateEmailAndPassword(req.Context(), database.UpdateEmailAndPasswordParams{
			ID:             userID,
			Email:          userReq.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			http.Error(w, "could not update user in db", http.StatusInternalServerError)
			return
		}

		user := apiUser{
			Email: dbUser.Email,
		}

		w.WriteHeader(http.StatusOK)
		writeResponse(user, w)
	}
}

type accessToken struct {
	Token string `json:"token"`
}

func HandlerRefresh(db authStore, secret string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(w, "could not get bearer token from header", http.StatusUnauthorized)
			return
		}

		refreshToken, err := db.GetRefreshToken(req.Context(), token)
		if err != nil {
			http.Error(w, "could not find token in db", http.StatusUnauthorized)
			return
		}

		if refreshToken.RevokedAt.Valid {
			http.Error(w, "token has been revoked", http.StatusUnauthorized)
			return
		}

		accessToken := accessToken{}
		accessToken.Token, err = auth.MakeJWT(refreshToken.UserID, secret)
		if err != nil {
			log.Printf("Error: could not make JWT: %v", err)
			http.Error(w, "could not make JWT", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		writeResponse(accessToken, w)
	}
}

func HandlerRevoke(db authStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(w, "could not get bearer token from header", http.StatusUnauthorized)
			return
		}

		if err = db.RevokeRefreshToken(req.Context(), token); err != nil {
			log.Printf("Error: could not revoke token: %v", err)
			http.Error(w, "could not revoke token", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type responseTypes interface {
	apiChirp | apiUser | []apiChirp | accessToken
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
