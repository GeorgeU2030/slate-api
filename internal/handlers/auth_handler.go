package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"slate-api/internal/auth"
	"slate-api/internal/config"
	"slate-api/internal/dto"
	"slate-api/internal/httputil"
	"slate-api/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, cfg: cfg}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body dto.UserRegisterRequest
	if !httputil.DecodeJSON(w, r, &body) {
		return
	}

	user, err := h.authService.Register(body.Email, body.Password, body.FullName)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyRegistered) {
			httputil.WriteError(w, http.StatusBadRequest, "Email already registered")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	respondWithTokens(w, r, h.authService, h.cfg, user, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body dto.UserLoginRequest
	if !httputil.DecodeJSON(w, r, &body) {
		return
	}

	user, err := h.authService.Login(body.Email, body.Password)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	respondWithTokens(w, r, h.authService, h.cfg, user, http.StatusOK)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	bodyToken := optionalBodyRefreshToken(r)

	incoming := auth.IncomingRefreshToken(r, bodyToken)
	if incoming == "" {
		httputil.WriteError(w, http.StatusUnauthorized, "Missing or invalid session")
		return
	}

	user, err := h.authService.Refresh(incoming)
	if err != nil {
		auth.ClearRefreshCookie(w, h.cfg.IsProduction())
		httputil.WriteError(w, http.StatusUnauthorized, "Missing or invalid session")
		return
	}

	respondWithTokens(w, r, h.authService, h.cfg, user, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	bodyToken := optionalBodyRefreshToken(r)
	incoming := auth.IncomingRefreshToken(r, bodyToken)

	_ = h.authService.Logout(incoming)

	auth.ClearRefreshCookie(w, h.cfg.IsProduction())
	w.WriteHeader(http.StatusNoContent)
}

func optionalBodyRefreshToken(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	var body dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return ""
	}
	if body.RefreshToken == nil {
		return ""
	}
	return *body.RefreshToken
}