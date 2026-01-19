package auth

import "testing"

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
