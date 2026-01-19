package public

import (
	"context"
	"errors"
	"time"

	"github.com/bailey4770/chirpy/internal/database"
	"github.com/google/uuid"
)

type mockDB struct {
	users  []database.User
	chirps []database.Chirp
}

func (m *mockDB) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.CreateUserRow, error) {
	returnUser := database.CreateUserRow{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     arg.Email,
	}

	savedUser := database.User{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          arg.Email,
		HashedPassword: arg.HashedPassword,
	}

	m.users = append(m.users, savedUser)
	return returnUser, nil
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

func (m *mockDB) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}

	return database.User{}, errors.New("could not find user in db")
}
