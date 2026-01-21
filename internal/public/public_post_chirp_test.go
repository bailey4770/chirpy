package public

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bailey4770/chirpy/internal/auth"
	"github.com/bailey4770/chirpy/internal/database"
	"github.com/google/uuid"
)

type mockChirpDB struct {
	chirps []database.Chirp
}

func (m *mockChirpDB) CreateChirp(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error) {
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

func TestHandleCreateChirp(t *testing.T) {
	type chirpTestCase struct {
		name               string
		params             chirpParams
		token              string
		expectedStatusCode int
		expectedBody       string
	}

	testCases := []chirpTestCase{
		{
			name: "Valid chirp",
			params: chirpParams{
				Body: "I had something interesting for breakfast",
			},
			expectedStatusCode: 201,
			expectedBody:       "I had something interesting for breakfast",
		},
		{
			name: "Too long chirp",
			params: chirpParams{
				Body: "lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			},
			expectedStatusCode: 400,
			expectedBody:       "",
		},
		{
			name: "Chirp with profanity",
			params: chirpParams{
				Body: "I hear Mastodon is better than Chirpy. sharbert I need to migrate",
			},
			expectedStatusCode: 201,
			expectedBody:       "I hear Mastodon is better than Chirpy. **** I need to migrate",
		},
		{
			name: "Chirp with profanity and uppercase",
			params: chirpParams{
				Body: "I really need a kerfuffle to go to bed sooner, Fornax !",
			},
			expectedStatusCode: 201,
			expectedBody:       "I really need a **** to go to bed sooner, **** !",
		},
		{
			name: "chirp with invalid token",
			params: chirpParams{
				Body: "I had something interesting for breakfast",
			},
			token:              "invalid.jwt.token",
			expectedStatusCode: 401,
		},
	}

	const url = "/api/chirps"
	const tokenSecret = "abcd"
	mock := &mockChirpDB{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.token == "" {
				tc.token, _ = auth.MakeJWT(uuid.New(), tokenSecret)
			}

			reqBody, _ := json.Marshal(&tc.params)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.token))

			w := httptest.NewRecorder()

			handler := HandlerPostChirp(mock, tokenSecret)
			handler(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()
			respBody, _ := io.ReadAll(resp.Body)

			if resp.StatusCode >= 400 {
				if resp.StatusCode != tc.expectedStatusCode {
					t.Fatalf("Fail: expected response status code %d but received %d with message: \n%s", tc.expectedStatusCode, resp.StatusCode, respBody)
				} else {
					return
				}
			}

			var response database.Chirp
			if err := json.Unmarshal(respBody, &response); err != nil {
				t.Fatalf("Error: could not unmarshal response: %v", err)
			}

			if response.Body != tc.expectedBody {
				t.Fatalf("Fail: expected cleaned body %s but recieved %s", tc.expectedBody, response.Body)
			}
		})
	}
}
