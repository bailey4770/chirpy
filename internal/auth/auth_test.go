package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashAndCheckPipeline(t *testing.T) {
	type testCase struct {
		testName       string
		plainPassword  string
		hashedPassword string
	}

	testCases := []testCase{
		{
			testName:      "pa$$word test",
			plainPassword: "pa$$word",
		},
		{
			testName:      "number password",
			plainPassword: "123456789",
		},
		{
			testName:      "white space test",
			plainPassword: " ",
		},
		{
			testName:      "long lorem ipsum test",
			plainPassword: "Lorem ipsum dolor sit amet consectetur adipiscing elit quisque faucibus ex sapien vitae pellentesque sem placerat in id cursus mi pretium tellus duis convallis tempus leo eu aenean sed diam urna tempor pulvinar vivamus fringilla lacus nec metus bibendum egestas iaculis massa nisl malesuada lacinia integer nunc posuere ut hendrerit.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			tc.hashedPassword, _ = HashPassword(tc.plainPassword)
			ok, _ := CheckPasswordHash(tc.plainPassword, tc.hashedPassword)
			if !ok {
				t.Fatalf("Fail: plain password %s hashed to %s but comparison check failed", tc.plainPassword, tc.hashedPassword)
			}
		})
	}
}

func TestJWTPipeline(t *testing.T) {
	type testCase struct {
		testName              string
		userID                uuid.UUID
		signingTokenSecret    string
		validatingTokenSecret string
		expiresIn             time.Duration
		expectedError         bool
	}

	testCases := []testCase{
		{
			testName:              "24h token",
			userID:                uuid.New(),
			signingTokenSecret:    "password",
			validatingTokenSecret: "password",
			expiresIn:             24 * time.Hour,
			expectedError:         false,
		},
		{
			testName:              "Immediate Expiration",
			userID:                uuid.New(),
			signingTokenSecret:    "testing",
			validatingTokenSecret: "testing",
			expiresIn:             0 * time.Second,
			expectedError:         true,
		},
		{
			testName:              "Wrong validating token secret",
			userID:                uuid.New(),
			signingTokenSecret:    "signing secret",
			validatingTokenSecret: "validating secret",
			expiresIn:             1 * time.Hour,
			expectedError:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			tokenString, _ := MakeJWT(tc.userID, tc.signingTokenSecret, tc.expiresIn)

			returnedUserID, err := ValidateJWT(tokenString, tc.validatingTokenSecret)
			if err != nil {
				if !tc.expectedError {
					t.Fatalf("Fail: unexpected error validating JWT: %v", err)
				}
			} else if tc.expectedError {
				t.Fatal("Fail: expected an error validating JWT but none returned")
			} else if returnedUserID != tc.userID {
				t.Fatalf("Fail: expected userID to be %v but received %v", tc.userID, returnedUserID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	type testCase struct {
		testName      string
		headers       http.Header
		expectedToken string
		expectedError bool
	}

	testCases := []testCase{
		{
			testName: "normal token in header",
			headers: http.Header{
				"Authorization": []string{"Bearer 123456789"},
				"Content-Type":  []string{"JSON"},
			},
			expectedToken: "123456789",
			expectedError: false,
		},
		{
			testName: "missing Bearer prefix",
			headers: http.Header{
				"Authorization": []string{"123456789"},
				"Content-Type":  []string{"JSON"},
			},
			expectedToken: "",
			expectedError: true,
		},
		{
			testName: "small but correct header",
			headers: http.Header{
				"Authorization": []string{"Bearer 123456789"},
			},
			expectedToken: "123456789",
			expectedError: false,
		},
		{
			testName: "no Authorization header provided",
			headers: http.Header{
				"Content-Type": []string{"JSON"},
			},
			expectedToken: "",
			expectedError: true,
		},
		{
			testName: "no token provided",
			headers: http.Header{
				"Authorization": []string{"Bearer "},
				"Content-Type":  []string{"JSON"},
			},
			expectedToken: "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			receivedToken, err := GetBearerToken(tc.headers)
			if err != nil && !tc.expectedError {
				t.Fatalf("Fail: expected no error but received: %v", err)
			} else if err == nil && tc.expectedError {
				t.Fatal("Fail: expected error but none received")
			}

			if receivedToken != tc.expectedToken {
				t.Fatalf("Fail: expected %s but received %s token", tc.expectedToken, receivedToken)
			}
		})
	}
}
