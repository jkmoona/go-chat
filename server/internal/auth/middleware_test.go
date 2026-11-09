package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	Setup("test-access-secret", "test-refresh-secret", false)
}

func TestAuthMiddlewareRejectsNoToken(t *testing.T) {
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "authentication required" {
		t.Fatalf("error = %q, want %q", body["error"], "authentication required")
	}
}

func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "garbage"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	token, err := GenerateAccessToken(42, "alice", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	var gotID string
	var gotUsername string
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = GetUserID(r.Context())
		gotUsername, _ = GetUsername(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotID != "42" {
		t.Fatalf("userID = %q, want %q", gotID, "42")
	}
	if gotUsername != "alice" {
		t.Fatalf("username = %q, want %q", gotUsername, "alice")
	}
}

func TestAuthMiddlewareRefreshFallback(t *testing.T) {
	refresh, err := GenerateRefreshToken(7, "bob", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	var gotID string
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refresh})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotID != "7" {
		t.Fatalf("userID = %q, want %q", gotID, "7")
	}

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "access_token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected new access_token cookie to be set")
	}
}

func TestAuthMiddlewareRejectsExpiredToken(t *testing.T) {
	token, err := GenerateAccessToken(1, "expired", -1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestOptionalAuthPassesWithoutToken(t *testing.T) {
	called := false
	handler := OptionalAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, hasID := GetUserID(r.Context())
		if hasID {
			t.Fatal("expected no user ID without token")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestOptionalAuthSetsContextWithToken(t *testing.T) {
	token, _ := GenerateAccessToken(99, "carol", 15*time.Minute)

	var gotID string
	handler := OptionalAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if gotID != "99" {
		t.Fatalf("userID = %q, want %q", gotID, "99")
	}
}

func TestGetUserIDIntFromContext(t *testing.T) {
	token, _ := GenerateAccessToken(55, "dave", 15*time.Minute)

	var gotInt int64
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotInt, _ = GetUserIDInt(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if gotInt != 55 {
		t.Fatalf("userIDInt = %d, want %d", gotInt, 55)
	}
}
