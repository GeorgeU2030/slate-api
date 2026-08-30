package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func RequireAuth(tokenService *TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := extractUserID(r, tokenService)
			if !ok {
				http.Error(w, `{"detail":"Missing or invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(tokenService *TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if userID, ok := extractUserID(r, tokenService); ok {
				ctx := context.WithValue(r.Context(), userIDContextKey, userID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractUserID(r *http.Request, tokenService *TokenService) (uuid.UUID, bool) {
	header := r.Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return uuid.Nil, false
	}

	tokenString := strings.TrimPrefix(header, "Bearer ")

	claims, err := tokenService.ParseAccessToken(tokenString)
	if err != nil {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}