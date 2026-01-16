package public

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleValidateChirp(t *testing.T) {
	type validateChirpTestCase struct {
		name                string
		params              chirpParams
		expectedStatusCode  int
		expectedError       string
		expectedValid       bool
		expectedCleanedBody string
	}
	testCases := []validateChirpTestCase{
		{
			name:                "Valid chirp",
			params:              chirpParams{Body: "I had something interesting for breakfast"},
			expectedStatusCode:  200,
			expectedError:       "",
			expectedValid:       true,
			expectedCleanedBody: "I had something interesting for breakfast",
		},
		{
			name:                "Too long chirp",
			params:              chirpParams{Body: "lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."},
			expectedStatusCode:  400,
			expectedError:       "Chirp is too long\n",
			expectedValid:       false,
			expectedCleanedBody: "",
		},
		{
			name:                "Chirp with profanity",
			params:              chirpParams{Body: "I hear Mastodon is better than Chirpy. sharbert I need to migrate"},
			expectedStatusCode:  200,
			expectedError:       "",
			expectedValid:       true,
			expectedCleanedBody: "I hear Mastodon is better than Chirpy. **** I need to migrate",
		},
		{
			name:                "Chirp with profanity and uppercase",
			params:              chirpParams{Body: "I really need a kerfuffle to go to bed sooner, Fornax !"},
			expectedStatusCode:  200,
			expectedError:       "",
			expectedValid:       true,
			expectedCleanedBody: "I really need a **** to go to bed sooner, **** !",
		},
	}

	const url = "/api/validate_chirp"

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(&testCase.params)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			HandlerValidateChirp(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != testCase.expectedStatusCode {
				t.Fatalf("Fail: expected response status code %d but received %d", resp.StatusCode, testCase.expectedStatusCode)
			}

			respBody, _ := io.ReadAll(resp.Body)

			if resp.StatusCode >= 400 {
				if string(respBody) != testCase.expectedError {
					t.Fatalf("Fail: expected response body %s but received %s", testCase.expectedError, string(respBody))
				}
				return
			}

			var response chirpReturnVals
			if err := json.Unmarshal(respBody, &response); err != nil {
				t.Fatalf("Error: could not unmarshal response: %v", err)
			}

			if response.Error != testCase.expectedError {
				t.Fatalf("Fail: expected error %s but recieved %s", testCase.expectedError, response.Error)
			}
			if response.Valid != testCase.expectedValid {
				t.Fatalf("Fail: expected valid %v but recieved %v", testCase.expectedValid, response.Valid)
			}
			if response.CleanedBody != testCase.expectedCleanedBody {
				t.Fatalf("Fail: expected cleaned body %s but recieved %s", testCase.expectedCleanedBody, response.CleanedBody)
			}
		})
	}
}
