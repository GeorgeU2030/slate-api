package httputil

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"net/http"
)

type ErrorResponse struct {
	Detail string `json:"detail"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func WriteError(w http.ResponseWriter, status int, detail string) {
	WriteJSON(w, status, ErrorResponse{Detail: detail})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := jsonv2.UnmarshalRead(r.Body, dst); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return false
	}
	return true
}