package public

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bailey4770/chirpy/internal/database"
)

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
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Error: could not unmarshal response: %v", err)
			}

			if testCase.params.Email != response.Email {
				t.Fatalf("Fail: expected email %s but recieved %s", testCase.params.Email, response.Email)
			}
		})
	}
}
