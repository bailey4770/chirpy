// Package auth provides functions for hashing a new password and checking if an inputted password matches a provided hash}
package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashed_password, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("could not hash passsword: %v", err)
	}

	return hashed_password, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	ok, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("could not compare password to saved hash: %v", err)
	}

	return ok, nil
}
