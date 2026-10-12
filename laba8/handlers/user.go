package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"laba8/models"
	"laba8/storage"
)

type UserHandler struct {
	store storage.UserStorage
}

func NewUserHandler(s storage.UserStorage) *UserHandler {
	return &UserHandler{store: s}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) error {
	f, err := parseFilter(r)
	if err != nil {
		return err
	}

	users, err := h.store.List(f)
	if err != nil {
		return Internal(err)
	}
	writeJSON(w, http.StatusOK, users)
	return nil
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return BadRequest("invalid id")
	}

	u, err := h.store.Get(id)
	if err != nil {
		return mapStorageError(err)
	}
	writeJSON(w, http.StatusOK, u)
	return nil
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return BadRequest("invalid JSON")
	}

	if err := u.Validate(); err != nil {
		return err
	}

	created, err := h.store.Create(u)
	if err != nil {
		return Internal(err)
	}
	writeJSON(w, http.StatusCreated, created)
	return nil
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return BadRequest("invalid id")
	}

	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return BadRequest("invalid JSON")
	}

	if err := u.Validate(); err != nil {
		return err
	}

	updated, err := h.store.Update(id, u)
	if err != nil {
		return mapStorageError(err)
	}
	writeJSON(w, http.StatusOK, updated)
	return nil
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return BadRequest("invalid id")
	}

	if err := h.store.Delete(id); err != nil {
		return mapStorageError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func parseFilter(r *http.Request) (storage.UserFilter, error) {
	q := r.URL.Query()
	var f storage.UserFilter

	f.Name = strings.TrimSpace(q.Get("name"))

	if v := q.Get("min_age"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return f, BadRequest("min_age must be a non-negative integer")
		}
		f.MinAge = n
	}
	if v := q.Get("max_age"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return f, BadRequest("max_age must be a non-negative integer")
		}
		f.MaxAge = n
	}
	if f.MinAge > 0 && f.MaxAge > 0 && f.MinAge > f.MaxAge {
		return f, BadRequest("min_age must be <= max_age")
	}

	// limit: по умолчанию 20, максимум 100 — иначе клиент положит сервер одним запросом.
	f.Limit = 20
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return f, BadRequest("limit must be a positive integer")
		}
		if n > 100 {
			return f, BadRequest("limit must be at most 100")
		}
		f.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return f, BadRequest("offset must be a non-negative integer")
		}
		f.Offset = n
	}

	return f, nil
}
