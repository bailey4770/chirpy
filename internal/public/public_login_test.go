package public

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bailey4770/chirpy/internal/auth"
	"github.com/bailey4770/chirpy/internal/database"
)

func TestLogin(t *testing.T) {
	type testCase struct {
		testName         string
		Email            string `json:"email"`
		registerPassword string
		LoginPassword    string `json:"password"`
		expectedError    bool
	}

	testCases := []testCase{
		{
			testName:         "standard login",
			Email:            "user@test.com",
			registerPassword: "pa$$word",
			LoginPassword:    "pa$$word",
			expectedError:    false,
		},
		{
			testName:         "false login",
			Email:            "user@test.com",
			registerPassword: "pa$$word",
			LoginPassword:    " ",
			expectedError:    true,
		},
	}

	const url = "/api/users"
	mock := &mockDB{}

	for _, tc := range testCases {
		hashedPassword, _ := auth.HashPassword(tc.registerPassword)

		t.Run(tc.testName, func(t *testing.T) {
			_, _ = mock.CreateUser(context.Background(), database.CreateUserParams{
				Email:          tc.Email,
				HashedPassword: hashedPassword,
			})

			reqBody, _ := json.Marshal(&tc)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := HandlerLogin(mock)
			handler(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if tc.expectedError && resp.StatusCode < 400 {
				t.Fatalf("Fail: expected to fail login but received %v status code", resp.StatusCode)
			} else if !tc.expectedError && resp.StatusCode >= 400 {
				t.Fatalf("Fail: expected to successfully log in but received %v status code", resp.StatusCode)
			}
		})
	}
}
