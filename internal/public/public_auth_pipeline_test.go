package public

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bailey4770/chirpy/internal/auth"
	"github.com/bailey4770/chirpy/internal/database"
	"github.com/google/uuid"
)

type mockAuthDB struct {
	users         []database.User
	refreshTokens []database.RefreshToken
}

// --- users ---

func (m *mockAuthDB) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return database.User{}, errors.New("user not found")
}

func (m *mockAuthDB) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.CreateUserRow, error) {
	user := database.User{
		ID:             uuid.New(),
		Email:          arg.Email,
		HashedPassword: arg.HashedPassword,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	m.users = append(m.users, user)

	return database.CreateUserRow{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// --- refresh tokens ---

func (m *mockAuthDB) CreateRefreshToken(
	ctx context.Context,
	arg database.CreateRefreshTokenParams,
) (database.RefreshToken, error) {
	token := database.RefreshToken{
		Token:     arg.Token,
		UserID:    arg.UserID,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.refreshTokens = append(m.refreshTokens, token)
	return token, nil
}

func (m *mockAuthDB) GetRefreshToken(ctx context.Context, token string) (database.RefreshToken, error) {
	for _, rt := range m.refreshTokens {
		if rt.Token == token {
			return rt, nil
		}
	}
	return database.RefreshToken{}, errors.New("token not found")
}

func (m *mockAuthDB) RevokeRefreshToken(ctx context.Context, token string) error {
	for i, rt := range m.refreshTokens {
		if rt.Token == token {
			rt.RevokedAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
			m.refreshTokens[i] = rt
			return nil
		}
	}
	return errors.New("token not found")
}

func TestAuthPipeline(t *testing.T) {
	const (
		email      = "user@test.com"
		password   = "pa$$word"
		secret     = "abcd"
		loginURL   = "/api/login"
		refreshURL = "/api/refresh"
		revokeURL  = "/api/revoke"
	)

	mock := &mockAuthDB{}

	// Seed user
	hashedPassword, _ := auth.HashPassword(password)
	_, _ = mock.CreateUser(context.Background(), database.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})

	// ---- LOGIN ----
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	loginReq := httptest.NewRequest(http.MethodPost, loginURL, bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()

	HandlerLogin(mock, secret)(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", loginRec.Code)
	}

	var loginResp apiUser
	_ = json.NewDecoder(loginRec.Body).Decode(&loginResp)

	if loginResp.RefreshToken == "" {
		t.Fatal("expected refresh token on login")
	}

	// ---- REFRESH (should succeed) ----
	refreshReq := httptest.NewRequest(http.MethodPost, refreshURL, nil)
	refreshReq.Header.Set("Authorization", "Bearer "+loginResp.RefreshToken)
	refreshRec := httptest.NewRecorder()

	HandlerRefresh(mock, secret)(refreshRec, refreshReq)

	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh failed: %d", refreshRec.Code)
	}

	// ---- REVOKE ----
	revokeReq := httptest.NewRequest(http.MethodPost, revokeURL, nil)
	revokeReq.Header.Set("Authorization", "Bearer "+loginResp.RefreshToken)
	revokeRec := httptest.NewRecorder()

	HandlerRevoke(mock)(revokeRec, revokeReq)

	if revokeRec.Code != http.StatusNoContent {
		t.Fatalf("revoke failed: %d", revokeRec.Code)
	}

	// ---- REFRESH AGAIN (should fail) ----
	refreshReq2 := httptest.NewRequest(http.MethodPost, refreshURL, nil)
	refreshReq2.Header.Set("Authorization", "Bearer "+loginResp.RefreshToken)
	refreshRec2 := httptest.NewRecorder()

	HandlerRefresh(mock, secret)(refreshRec2, refreshReq2)

	if refreshRec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked token to fail, got %d", refreshRec2.Code)
	}
}
