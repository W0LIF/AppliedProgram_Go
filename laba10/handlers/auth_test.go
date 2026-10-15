package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"laba10/models"
	"laba10/storage"
)

func newAuthServer() (*http.ServeMux, *storage.MemoryStorage, *storage.TokenStore) {
	store := storage.NewMemoryStorage()
	tokens := storage.NewTokenStore()
	h := NewUserHandler(store)
	a := NewAuthHandler(store, tokens)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", Wrap(a.Register))
	mux.HandleFunc("POST /auth/login", Wrap(a.Login))
	mux.HandleFunc("POST /auth/logout", Wrap(a.Logout))
	mux.HandleFunc("GET /users", Wrap(RequireAuth(tokens, h.List)))
	return mux, store, tokens
}

func postJSON(mux *http.ServeMux, path, token string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest("POST", path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestRegister_OK(t *testing.T) {
	mux, _, _ := newAuthServer()
	rr := postJSON(mux, "/auth/register", "", models.RegisterRequest{
		Name: "Alice", Email: "a@x.com", Age: 25, Password: "secret1",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body)
	}
	var resp models.AuthResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Token == "" || resp.User.ID == 0 {
		t.Errorf("unexpected body: %+v", resp)
	}
}

func TestRegister_400_WeakPassword(t *testing.T) {
	mux, _, _ := newAuthServer()
	rr := postJSON(mux, "/auth/register", "", models.RegisterRequest{
		Name: "Alice", Email: "a@x.com", Age: 25, Password: "123",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "password") {
		t.Errorf("details should mention password: %s", rr.Body)
	}
}

func TestRegister_400_DuplicateEmail(t *testing.T) {
	mux, _, _ := newAuthServer()
	body := models.RegisterRequest{Name: "A", Email: "a@x.com", Age: 20, Password: "secret1"}
	_ = postJSON(mux, "/auth/register", "", body)
	rr := postJSON(mux, "/auth/register", "", body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestLogin_OK(t *testing.T) {
	mux, _, _ := newAuthServer()
	_ = postJSON(mux, "/auth/register", "", models.RegisterRequest{
		Name: "Alice", Email: "a@x.com", Age: 25, Password: "secret1",
	})
	rr := postJSON(mux, "/auth/login", "", models.LoginRequest{
		Email: "a@x.com", Password: "secret1",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestLogin_401_WrongPassword(t *testing.T) {
	mux, _, _ := newAuthServer()
	_ = postJSON(mux, "/auth/register", "", models.RegisterRequest{
		Name: "Alice", Email: "a@x.com", Age: 25, Password: "secret1",
	})
	rr := postJSON(mux, "/auth/login", "", models.LoginRequest{
		Email: "a@x.com", Password: "wrong",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestLogin_401_UnknownUser(t *testing.T) {
	mux, _, _ := newAuthServer()
	rr := postJSON(mux, "/auth/login", "", models.LoginRequest{
		Email: "nobody@x.com", Password: "whatever",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_401_NoToken(t *testing.T) {
	mux, _, _ := newAuthServer()
	req := httptest.NewRequest("GET", "/users", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_401_BadToken(t *testing.T) {
	mux, _, _ := newAuthServer()
	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer deadbeef")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_200_WithToken(t *testing.T) {
	mux, store, tokens := newAuthServer()
	u, _ := store.CreateWithPassword(models.User{Name: "A", Email: "a@x.com", Age: 20}, "hash")
	token := tokens.Issue(u.ID)

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
}
