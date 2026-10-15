package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"laba10/auth"
	"laba10/models"
	"laba10/storage"
)

type AuthHandler struct {
	store    storage.UserStorage
	tokens   *auth.TokenManager
	sessions *auth.SessionStore
}

func NewAuthHandler(s storage.UserStorage, tm *auth.TokenManager, ss *auth.SessionStore) *AuthHandler {
	return &AuthHandler{store: s, tokens: tm, sessions: ss}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return BadRequest("invalid JSON")
	}
	if err := req.Validate(); err != nil {
		return err
	}

	if _, _, err := h.store.FindByEmail(req.Email); err == nil {
		return BadRequest("email already registered")
	} else if !errors.Is(err, storage.ErrNotFound) {
		return Internal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return Internal(err)
	}

	u := models.User{Name: req.Name, Email: req.Email, Age: req.Age, Role: "user"}
	created, err := h.store.CreateWithPassword(u, string(hash))
	if err != nil {
		return Internal(err)
	}

	token, err := h.tokens.Issue(created.ID, created.Role)
	if err != nil {
		return Internal(err)
	}
	writeJSON(w, http.StatusCreated, models.AuthResponse{Token: token, User: created})
	return nil
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return BadRequest("invalid JSON")
	}
	if req.Email == "" || req.Password == "" {
		return BadRequest("email and password are required")
	}

	u, hash, err := h.store.FindByEmail(req.Email)
	if errors.Is(err, storage.ErrNotFound) {
		return Unauthorized("invalid credentials")
	}
	if err != nil {
		return Internal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return Unauthorized("invalid credentials")
	}

	token, err := h.tokens.Issue(u.ID, u.Role)
	if err != nil {
		return Internal(err)
	}
	writeJSON(w, http.StatusOK, models.AuthResponse{Token: token, User: u})
	return nil
}

func (h *AuthHandler) SessionLogin(w http.ResponseWriter, r *http.Request) error {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return BadRequest("invalid JSON")
	}

	u, hash, err := h.store.FindByEmail(req.Email)
	if errors.Is(err, storage.ErrNotFound) {
		return Unauthorized("invalid credentials")
	}
	if err != nil {
		return Internal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return Unauthorized("invalid credentials")
	}

	sessionID := h.sessions.Create(u.ID, u.Role)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	csrf := auth.NewCSRFToken()
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrf,
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"user":       u,
		"csrf_token": csrf,
	})
	return nil
}

func (h *AuthHandler) SessionLogout(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		h.sessions.Delete(cookie.Value)
	}
	for _, name := range []string{sessionCookieName, csrfCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: name == sessionCookieName,
			Secure:   true,
		})
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
