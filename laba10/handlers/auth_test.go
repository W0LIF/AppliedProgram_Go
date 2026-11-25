package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"laba10/auth"
	"laba10/models"
	"laba10/storage"
)

func newAuthServer() (*http.ServeMux, *storage.MemoryStorage, *auth.TokenManager) {
	store := storage.NewMemoryStorage()
	tm := auth.NewTokenManager("test-secret")
	ss := auth.NewSessionStore()
	h := NewUserHandler(store)
	a := NewAuthHandler(store, tm, ss)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", Wrap(a.Register))
	mux.HandleFunc("POST /auth/login", Wrap(a.Login))
	mux.HandleFunc("POST /auth/session-login", Wrap(a.SessionLogin))
	mux.HandleFunc("POST /auth/session-logout", Wrap(a.SessionLogout))
	mux.HandleFunc("GET /users", Wrap(RequireAuth(tm, ss, h.List)))
	mux.HandleFunc("POST /users", Wrap(RequireAuth(tm, ss, h.Create)))
	mux.HandleFunc("GET /admin/users", Wrap(
		RequireAuth(tm, ss, RequireRole(h.List, "admin"))))
	return mux, store, tm
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
	mux, store, tm := newAuthServer()
	u, _ := store.CreateWithPassword(models.User{Name: "A", Email: "a@x.com", Age: 20}, "hash")
	token, err := tm.Issue(u.ID, u.Role)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
}

func sessionLogin(t *testing.T, mux *http.ServeMux, email, password string) (string, string) {
	t.Helper()

	rr := postJSON(mux, "/auth/session-login", "", models.LoginRequest{
		Email: email, Password: password,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("session login: expected 200, got %d: %s", rr.Code, rr.Body)
	}

	var sessionID, csrf string
	for _, c := range rr.Result().Cookies() {
		switch c.Name {
		case sessionCookieName:
			sessionID = c.Value
		case csrfCookieName:
			csrf = c.Value
		}
	}
	if sessionID == "" || csrf == "" {
		t.Fatalf("session cookies not set: %v", rr.Result().Cookies())
	}
	return sessionID, csrf
}

func TestSessionLogin_200_AndAccess(t *testing.T) {
	mux, _, _ := newAuthServer()
	_ = postJSON(mux, "/auth/register", "", models.RegisterRequest{
		Name: "Alice", Email: "a@x.com", Age: 25, Password: "secret1",
	})

	sessionID, _ := sessionLogin(t, mux, "a@x.com", "secret1")

	req := httptest.NewRequest("GET", "/users", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
}

func TestSession_POST_RequiresCSRF(t *testing.T) {
	mux, _, _ := newAuthServer()
	_ = postJSON(mux, "/auth/register", "", models.RegisterRequest{
		Name: "Alice", Email: "a@x.com", Age: 25, Password: "secret1",
	})
	sessionID, csrf := sessionLogin(t, mux, "a@x.com", "secret1")

	body, _ := json.Marshal(models.User{Name: "Bob", Email: "b@x.com", Age: 30})

	// Без CSRF-заголовка — 403.
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrf})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("without csrf: expected 403, got %d", rr.Code)
	}

	// Неверный CSRF — 403.
	req = httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(csrfHeaderName, "wrong-token")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrf})
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("wrong csrf: expected 403, got %d", rr.Code)
	}

	// С корректным CSRF — 201.
	req = httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(csrfHeaderName, csrf)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrf})
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("with csrf: expected 201, got %d: %s", rr.Code, rr.Body)
	}
}

func TestSessionLogout_InvalidatesSession(t *testing.T) {
	mux, _, _ := newAuthServer()
	_ = postJSON(mux, "/auth/register", "", models.RegisterRequest{
		Name: "Alice", Email: "a@x.com", Age: 25, Password: "secret1",
	})
	sessionID, _ := sessionLogin(t, mux, "a@x.com", "secret1")

	req := httptest.NewRequest("POST", "/auth/session-logout", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("logout: expected 204, got %d", rr.Code)
	}

	req = httptest.NewRequest("GET", "/users", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("after logout: expected 401, got %d", rr.Code)
	}
}

func TestRequireRole(t *testing.T) {
	mux, store, tm := newAuthServer()

	admin, _ := store.CreateWithPassword(
		models.User{Name: "Admin", Email: "admin@x.com", Age: 30, Role: "admin"}, "hash")
	adminToken, _ := tm.Issue(admin.ID, admin.Role)

	user, _ := store.CreateWithPassword(
		models.User{Name: "User", Email: "u@x.com", Age: 20, Role: "user"}, "hash")
	userToken, _ := tm.Issue(user.ID, user.Role)

	get := func(token string) int {
		req := httptest.NewRequest("GET", "/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		return rr.Code
	}

	if code := get(userToken); code != http.StatusForbidden {
		t.Errorf("user role: expected 403, got %d", code)
	}
	if code := get(adminToken); code != http.StatusOK {
		t.Errorf("admin role: expected 200, got %d", code)
	}
}
