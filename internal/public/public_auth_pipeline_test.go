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

// --- integration test ---

func TestAuthPipeline(t *testing.T) {
	const (
		email       = "user@test.com"
		password    = "pa$$word"
		newEmail    = "new@test.com"
		newPassword = "newpa$$word"
		secret      = "abcd"

		loginURL   = "/api/login"
		updateURL  = "/api/users"
		refreshURL = "/api/refresh"
		revokeURL  = "/api/revoke"
	)

	ctx := authTestCtx{
		t:      t,
		db:     &mockAuthDB{},
		secret: secret,
	}

	seedUser(ctx, email, password)

	loginResp := login(ctx, email, password, loginURL, http.StatusOK)
	updateUser(ctx, loginResp.Token, newEmail, newPassword, updateURL)

	login(ctx, newEmail, password, loginURL, http.StatusUnauthorized)
	newLogin := login(ctx, newEmail, newPassword, loginURL, http.StatusOK)

	refresh(ctx, newLogin.RefreshToken, refreshURL, http.StatusOK)
	revoke(ctx, newLogin.RefreshToken, revokeURL)
	refresh(ctx, newLogin.RefreshToken, refreshURL, http.StatusUnauthorized)
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

func (m *mockAuthDB) UpdateEmailAndPassword(
	ctx context.Context,
	arg database.UpdateEmailAndPasswordParams,
) (database.UpdateEmailAndPasswordRow, error) {
	for i, u := range m.users {
		if u.ID == arg.ID {
			u.Email = arg.Email
			u.HashedPassword = arg.HashedPassword
			u.UpdatedAt = time.Now()
			m.users[i] = u
			return database.UpdateEmailAndPasswordRow{Email: arg.Email}, nil
		}
	}
	return database.UpdateEmailAndPasswordRow{}, errors.New("user not found")
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

// --- test helpers ---

type authTestCtx struct {
	t      *testing.T
	db     *mockAuthDB
	secret string
}

func seedUser(ctx authTestCtx, email, password string) {
	hashed, _ := auth.HashPassword(password)
	_, _ = ctx.db.CreateUser(context.Background(), database.CreateUserParams{
		Email:          email,
		HashedPassword: hashed,
	})
}

func login(
	ctx authTestCtx,
	email, password, url string,
	expectStatus int,
) apiUser {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	HandlerLogin(ctx.db, ctx.secret)(rec, req)

	if rec.Code != expectStatus {
		ctx.t.Fatalf("login expected %d, got %d", expectStatus, rec.Code)
	}

	if expectStatus != http.StatusOK {
		return apiUser{}
	}

	var resp apiUser
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	return resp
}

func updateUser(
	ctx authTestCtx,
	token, email, password, url string,
) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	HandlerUpdateEmailAndPassword(ctx.db, ctx.secret)(rec, req)

	if rec.Code != http.StatusOK {
		ctx.t.Fatalf("update failed: %d", rec.Code)
	}
}

func refresh(
	ctx authTestCtx,
	refreshToken, url string,
	expectStatus int,
) {
	req := httptest.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()

	HandlerRefresh(ctx.db, ctx.secret)(rec, req)

	if rec.Code != expectStatus {
		ctx.t.Fatalf("refresh expected %d, got %d", expectStatus, rec.Code)
	}
}

func revoke(
	ctx authTestCtx,
	refreshToken, url string,
) {
	req := httptest.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()

	HandlerRevoke(ctx.db)(rec, req)

	if rec.Code != http.StatusNoContent {
		ctx.t.Fatalf("revoke failed: %d", rec.Code)
	}
}
