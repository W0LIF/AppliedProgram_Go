package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"laba9/models"
	"laba9/storage"
)

type AuthHandler struct {
	store  storage.UserStorage
	tokens *storage.TokenStore
}

func NewAuthHandler(s storage.UserStorage, t *storage.TokenStore) *AuthHandler {
	return &AuthHandler{store: s, tokens: t}
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

	u := models.User{Name: req.Name, Email: req.Email, Age: req.Age}
	created, err := h.store.CreateWithPassword(u, string(hash))
	if err != nil {
		return Internal(err)
	}

	token := h.tokens.Issue(created.ID)
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

	token := h.tokens.Issue(u.ID)
	writeJSON(w, http.StatusOK, models.AuthResponse{Token: token, User: u})
	return nil
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	if token := extractToken(r); token != "" {
		h.tokens.Revoke(token)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
