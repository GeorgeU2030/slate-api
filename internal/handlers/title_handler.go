package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"slate-api/internal/auth"
	"slate-api/internal/dto"
	"slate-api/internal/httputil"
	"slate-api/internal/models"
	"slate-api/internal/services"
)

type TitleHandler struct {
	titleService *services.TitleService
}

func NewTitleHandler(titleService *services.TitleService) *TitleHandler {
	return &TitleHandler{titleService: titleService}
}

func (h *TitleHandler) TrendingMovies(w http.ResponseWriter, r *http.Request) {
	result, err := h.titleService.GetTrending(r.Context(), models.TitleTypeMovie)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "TMDB trending movies failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *TitleHandler) TrendingTV(w http.ResponseWriter, r *http.Request) {
	result, err := h.titleService.GetTrending(r.Context(), models.TitleTypeTV)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "TMDB trending TV failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *TitleHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Query cannot be empty")
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Missing or invalid token")
		return
	}

	result, err := h.titleService.Search(r.Context(), &userID, q)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "TMDB search failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *TitleHandler) GetMyTitles(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Missing or invalid token")
		return
	}

	query := r.URL.Query()

	var year *int
	if v := query.Get("year"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "Invalid year")
			return
		}
		year = &parsed
	}

	var titleType *models.TitleType
	if v := query.Get("type"); v != "" {
		t := models.TitleType(v)
		titleType = &t
	}

	var search *string
	if v := query.Get("q"); v != "" {
		search = &v
	}

	result, err := h.titleService.GetMyTitles(userID, year, titleType, search)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to load titles")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *TitleHandler) RateTitle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Missing or invalid token")
		return
	}

	var body dto.RateTitleRequest
	if !httputil.DecodeJSON(w, r, &body) {
		return
	}
	if err := body.Validate(); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	result, err := h.titleService.Rate(r.Context(), userID, body)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "TMDB fetch failed")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, result)
}

func (h *TitleHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Missing or invalid token")
		return
	}

	userTitleID, err := uuid.Parse(chi.URLParam(r, "userTitleID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	result, err := h.titleService.GetDetail(userID, userTitleID)
	if err != nil {
		if errors.Is(err, services.ErrTitleNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "Rating not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to load title")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *TitleHandler) UpdateScore(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Missing or invalid token")
		return
	}

	userTitleID, err := uuid.Parse(chi.URLParam(r, "userTitleID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	var body dto.UpdateScoreRequest
	if !httputil.DecodeJSON(w, r, &body) {
		return
	}
	if err := body.Validate(); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	result, err := h.titleService.UpdateScore(userID, userTitleID, body)
	if err != nil {
		if errors.Is(err, services.ErrTitleNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "Rating not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to update rating")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *TitleHandler) DeleteRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Missing or invalid token")
		return
	}

	userTitleID, err := uuid.Parse(chi.URLParam(r, "userTitleID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	if err := h.titleService.DeleteRating(userID, userTitleID); err != nil {
		if errors.Is(err, services.ErrTitleNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "Rating not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to delete rating")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}