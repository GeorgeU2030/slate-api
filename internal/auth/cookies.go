package auth

import (
	"net/http"
	"time"
)

const RefreshCookieName = "refresh_token"

func SetRefreshCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearRefreshCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func IncomingRefreshToken(r *http.Request, bodyToken string) string {
	cookie, err := r.Cookie(RefreshCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return bodyToken
}