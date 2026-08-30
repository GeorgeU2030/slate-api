package handlers

import (
	"net/http"

	"slate-api/internal/auth"
	"slate-api/internal/config"
	"slate-api/internal/dto"
	"slate-api/internal/httputil"
	"slate-api/internal/models"
	"slate-api/internal/services"
)

func respondWithTokens(
	w http.ResponseWriter,
	r *http.Request,
	authService *services.AuthService,
	cfg *config.Config,
	user *models.User,
	status int,
) {
	accessToken, rawRefresh, expiresAt, err := authService.IssueTokens(user)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to issue tokens")
		return
	}

	if auth.GetClientType(r) == auth.ClientMobile {
		httputil.WriteJSON(w, status, dto.TokenPairResponse{
			AccessToken:  accessToken,
			RefreshToken: &rawRefresh,
		})
		return
	}

	auth.SetRefreshCookie(w, rawRefresh, expiresAt, cfg.IsProduction())
	httputil.WriteJSON(w, status, dto.TokenPairResponse{AccessToken: accessToken})
}