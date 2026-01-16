package public

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bailey4770/chirpy/internal/database"
	"github.com/google/uuid"
)

type mockDB struct {
	users []database.User
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

func TestHandleCreateUser(t *testing.T) {
	type createUserTestCase struct {
		name               string
		params             createUserParams
		expectedStatusCode int
	}

	testCases := []createUserTestCase{
		{
			name:               "Valid email 1",
			params:             createUserParams{Email: "mloneusk@example.co"},
			expectedStatusCode: 201,
		},
		{
			name:               "Valid email 2",
			params:             createUserParams{Email: "dackjorsey@example.co"},
			expectedStatusCode: 201,
		},
	}

	const url = "/api/users"
	mock := &mockDB{}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(&testCase.params)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := HandlerCreateUser(mock)
			handler(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if testCase.expectedStatusCode != resp.StatusCode {
				t.Fatalf("Fail: expected response status code %d but received %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			var response database.User
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&response); err != nil {
				t.Fatalf("Error: could not unmarshal response: %v", err)
			}

			if testCase.params.Email != response.Email {
				t.Fatalf("Fail: expected email %s but recieved %s", testCase.params.Email, response.Email)
			}
		})
	}
}
