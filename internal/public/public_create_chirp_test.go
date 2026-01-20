package public

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bailey4770/chirpy/internal/auth"
	"github.com/bailey4770/chirpy/internal/database"
	"github.com/google/uuid"
)

func TestHandleCreateChirp(t *testing.T) {
	type validateChirpTestCase struct {
		name               string
		params             chirpParams
		token              string
		expectedStatusCode int
		expectedBody       string
	}
	testCases := []validateChirpTestCase{
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
			token:              "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJjaGlycHkiLCJzdWIiOiIyOWViYTcxYS1mN2Y1LTRhYWYtYTFiNi01ZmFmNTIyY2ExNGYiLCJleHAiOjE3Njg5MjU1MTEsImlhdCI6MTc2ODkyMTkxMX0.UNMejxVMCRl1V0fGXCfNDYEP61c-1JQjCp17_9Mh2NE",
			expectedStatusCode: 401,
		},
	}

	const url = "/api/chirps"
	const tokenSecret = "abcd"
	mock := &mockDB{}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.token == "" {
				testCase.token, _ = auth.MakeJWT(uuid.New(), tokenSecret)
			}

			reqBody, _ := json.Marshal(&testCase.params)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testCase.token))

			w := httptest.NewRecorder()

			handler := HandlerPostChirp(mock, tokenSecret)
			handler(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()
			respBody, _ := io.ReadAll(resp.Body)

			if resp.StatusCode >= 400 {
				if resp.StatusCode != testCase.expectedStatusCode {
					t.Fatalf("Fail: expected response status code %d but received %d with message: \n%s", testCase.expectedStatusCode, resp.StatusCode, respBody)
				} else {
					return
				}
			}

			var response database.Chirp
			if err := json.Unmarshal(respBody, &response); err != nil {
				t.Fatalf("Error: could not unmarshal response: %v", err)
			}

			if response.Body != testCase.expectedBody {
				t.Fatalf("Fail: expected cleaned body %s but recieved %s", testCase.expectedBody, response.Body)
			}
		})
	}
}
