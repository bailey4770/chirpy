package public

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestHandlerValidateJSON(t *testing.T) {
	type testCase struct {
		name                string
		params              params
		expectedStatusCode  int
		expectedError       string
		expectedValid       bool
		expectedCleanedBody string
	}

	testCases := []testCase{
		{
			name:                "Valid chirp",
			params:              params{Body: "I had something interesting for breakfast"},
			expectedStatusCode:  200,
			expectedError:       "",
			expectedValid:       true,
			expectedCleanedBody: "I had something interesting for breakfast",
		},
		{
			name:                "Too long chirp",
			params:              params{Body: "lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."},
			expectedStatusCode:  400,
			expectedError:       "Chirp is too long",
			expectedValid:       false,
			expectedCleanedBody: "",
		},
		{
			name:                "Chirp with profanity",
			params:              params{Body: "I hear Mastodon is better than Chirpy. sharbert I need to migrate"},
			expectedStatusCode:  200,
			expectedError:       "",
			expectedValid:       true,
			expectedCleanedBody: "I hear Mastodon is better than Chirpy. **** I need to migrate",
		},
		{
			name:                "Chirp with profanity and uppercase",
			params:              params{Body: "I really need a kerfuffle to go to bed sooner, Fornax !"},
			expectedStatusCode:  200,
			expectedError:       "",
			expectedValid:       true,
			expectedCleanedBody: "I really need a **** to go to bed sooner, **** !",
		},
	}

	const url = "http://localhost:8080/api/validate_chirp"

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(&testCase.params)
			if err != nil {
				t.Fatalf("Error: could not marshal testcase %v: %v", testCase, err)
			}

			req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
			if err != nil {
				t.Fatalf("Error: could not make new request with body %s: %v", reqBody, err)
			}

			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error: http client could not DO request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("Error: could not close resp body: %v", err)
				}
			}()

			if resp.StatusCode != testCase.expectedStatusCode {
				t.Fatalf("Fail: expected response status code %d but received %d", resp.StatusCode, testCase.expectedStatusCode)
			}

			var response returnVals
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&response); err != nil {
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
