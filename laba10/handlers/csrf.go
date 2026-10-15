package handlers

import (
	"crypto/subtle"
	"net/http"

	"laba10/auth"
)

func EnsureCSRFCookie(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if _, err := r.Cookie(csrfCookieName); err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookieName,
				Value:    auth.NewCSRFToken(),
				Path:     "/",
				HttpOnly: false,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			})
		}
		return next(w, r)
	}
}

func checkCSRF(r *http.Request) error {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return Forbidden("csrf cookie missing")
	}
	header := r.Header.Get(csrfHeaderName)
	if header == "" {
		return Forbidden("csrf header missing")
	}
	if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
		return Forbidden("csrf token mismatch")
	}
	return nil
}
