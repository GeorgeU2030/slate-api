package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"slate-api/internal/auth"
	"slate-api/internal/httputil"
)

func registerRootRoutes(router chi.Router, db *gorm.DB, tokenService *auth.TokenService) {
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Polarky Slate API, is Live!"})
	})

	router.With(auth.RequireAuth(tokenService)).Get("/me", func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"userId": userID.String()})
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := db.DB()
		if err != nil {
			httputil.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			httputil.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
			return
		}

		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"name":      "Slate API",
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
}