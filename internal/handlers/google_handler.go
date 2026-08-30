package handlers

import (
	"net/http"

	"slate-api/internal/auth"
	"slate-api/internal/config"
	"slate-api/internal/dto"
	"slate-api/internal/httputil"
	"slate-api/internal/services"
)

type GoogleHandler struct {
	authService *services.AuthService
	googleAuth  *auth.GoogleAuthService
	cfg         *config.Config
}

func NewGoogleHandler(authService *services.AuthService, googleAuth *auth.GoogleAuthService, cfg *config.Config) *GoogleHandler {
	return &GoogleHandler{authService: authService, googleAuth: googleAuth, cfg: cfg}
}

func (h *GoogleHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var body dto.GoogleLoginRequest
	if !httputil.DecodeJSON(w, r, &body) {
		return
	}

	idToken, err := h.googleAuth.ExchangeCodeForIDToken(r.Context(), body.Code)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "Invalid Google authorization code")
		return
	}

	payload, err := h.googleAuth.ValidateIDToken(r.Context(), idToken, h.cfg.GoogleClientID)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "Invalid Google token")
		return
	}

	if !payload.EmailVerified {
		httputil.WriteError(w, http.StatusUnauthorized, "Email not verified by Google")
		return
	}

	user, err := h.authService.GetOrCreateGoogleUser(payload.Email, payload.Name, payload.Subject)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to process Google login")
		return
	}

	respondWithTokens(w, r, h.authService, h.cfg, user, http.StatusOK)
}