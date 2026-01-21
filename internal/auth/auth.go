// Package auth provides functions for hashing a new password and checking if an inputted password matches a provided hash}
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("could not hash passsword: %v", err)
	}

	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	ok, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("could not compare password to saved hash: %v", err)
	}

	return ok, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
		Subject:   userID.String(),
	})

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("could not sign token with secret: %v", err)
	}

	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(tokenSecret), nil
		})

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not parse token: %v", err)
	} else if !token.Valid {
		return uuid.UUID{}, errors.New("token is invalid")
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("could not parse subject field to UUID: %v", err)
		}
		return userID, nil
	}

	return uuid.UUID{}, errors.New("unknown claims type, cannot proceed")
}

func GetBearerToken(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("authorizaton header missing")
	}

	parts := strings.Fields(val)

	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid value set for Authorizaton header: %v", parts)
	}

	return parts[1], nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return hex.EncodeToString(key)
}

func GetAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("authorizaton header missing")
	}

	parts := strings.Fields(val)

	if len(parts) != 2 || parts[0] != "ApiKey" {
		return "", fmt.Errorf("invalid value set for Authorizaton header: %v", parts)
	}

	return parts[1], nil
}
