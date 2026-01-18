package public

import (
	"context"
	"time"

	"github.com/bailey4770/chirpy/internal/database"
	"github.com/google/uuid"
)

type mockDB struct {
	users  []database.User
	chirps []database.Chirp
}

func (m *mockDB) CreateUser(ctx context.Context, email string) (database.User, error) {
	user := database.User{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     email,
	}
	m.users = append(m.users, user)
	return user, nil
}

func (m *mockDB) CreateChirp(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error) {
	chirp := database.Chirp{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      arg.Body,
		UserID:    arg.UserID,
	}
	m.chirps = append(m.chirps, chirp)
	return chirp, nil
}
