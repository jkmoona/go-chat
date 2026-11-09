package auth

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/jkmoona/go-chat/server/internal/httputil"
)

type contextKey int

const (
	userIDKey contextKey = iota
	userIDIntKey
	usernameKey
)

func withUserContext(ctx context.Context, idStr, username string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, idStr)
	ctx = context.WithValue(ctx, usernameKey, username)
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		ctx = context.WithValue(ctx, userIDIntKey, id)
	}
	return ctx
}

func extractAuth(w http.ResponseWriter, r *http.Request) (*http.Request, bool) {
	tokenStr, err := httputil.ReadCookie(r, "access_token")
	if err == nil {
		claims, err := ValidateAccessToken(tokenStr)
		if err == nil {
			ctx := withUserContext(r.Context(), claims.ID, claims.Username)
			return r.WithContext(ctx), true
		}
	}

	refreshTokenStr, err := httputil.ReadCookie(r, "refresh_token")
	if err != nil {
		return r, false
	}

	claims, err := ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return r, false
	}

	userID, _ := strconv.ParseInt(claims.ID, 10, 64)
	newAccessToken, err := GenerateAccessToken(userID, claims.Username, 15*time.Minute)
	if err != nil {
		return r, false
	}

	httputil.SetCookie(w, "access_token", newAccessToken, 900, SecureCookies(), true)
	ctx := withUserContext(r.Context(), claims.ID, claims.Username)
	return r.WithContext(ctx), true
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2, ok := extractAuth(w, r)
		if !ok {
			httputil.ErrorJSON(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r2)
	})
}

func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2, _ := extractAuth(w, r)
		next.ServeHTTP(w, r2)
	})
}

func GetUserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

func GetUserIDInt(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(userIDIntKey).(int64)
	return v, ok
}

func GetUsername(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(usernameKey).(string)
	return v, ok
}
