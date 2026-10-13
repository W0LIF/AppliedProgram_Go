package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"laba9/models"
	"laba9/storage"
)

func newTestMux() (*http.ServeMux, *storage.MemoryStorage, string) {
	store := storage.NewMemoryStorage()
	tokens := storage.NewTokenStore()
	h := NewUserHandler(store)

	u, _ := store.CreateWithPassword(
		models.User{Name: "Tester", Email: "t@x.com", Age: 20},
		"hash",
	)
	token := tokens.Issue(u.ID)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", Wrap(RequireAuth(tokens, h.List)))
	mux.HandleFunc("GET /users/{id}", Wrap(RequireAuth(tokens, h.Get)))
	mux.HandleFunc("POST /users", Wrap(RequireAuth(tokens, h.Create)))
	mux.HandleFunc("PUT /users/{id}", Wrap(RequireAuth(tokens, h.Update)))
	mux.HandleFunc("DELETE /users/{id}", Wrap(RequireAuth(tokens, h.Delete)))
	return mux, store, token
}

func doJSON(mux *http.ServeMux, method, path, token string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestCreate_201(t *testing.T) {
	mux, _, token := newTestMux()

	rr := doJSON(mux, "POST", "/users", token,
		models.User{Name: "Alice", Email: "a@x.com", Age: 25})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body)
	}

	var got models.User
	_ = json.Unmarshal(rr.Body.Bytes(), &got)
	if got.Name != "Alice" {
		t.Errorf("unexpected body: %+v", got)
	}
	if got.ID <= 0 {
		t.Errorf("expected positive ID, got %d", got.ID)
	}
}

func TestCreate_400_InvalidJSON(t *testing.T) {
	mux, _, token := newTestMux()

	req := httptest.NewRequest("POST", "/users", strings.NewReader("not-json"))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreate_400_Validation(t *testing.T) {
	mux, _, token := newTestMux()

	rr := doJSON(mux, "POST", "/users", token,
		models.User{Name: "", Email: "x", Age: -1})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["error"] != "validation failed" {
		t.Errorf("expected validation failed, got %v", body["error"])
	}
	if _, ok := body["details"]; !ok {
		t.Errorf("details missing in %v", body)
	}
}

func TestCreate_401_NoToken(t *testing.T) {
	mux, _, _ := newTestMux()

	rr := doJSON(mux, "POST", "/users", "",
		models.User{Name: "Alice", Email: "a@x.com", Age: 25})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGet_200(t *testing.T) {
	mux, store, token := newTestMux()

	u, _ := store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "GET", "/users/"+itoa(u.ID), token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
}

func TestGet_404(t *testing.T) {
	mux, _, token := newTestMux()

	rr := doJSON(mux, "GET", "/users/999", token, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGet_400_InvalidID(t *testing.T) {
	mux, _, token := newTestMux()

	rr := doJSON(mux, "GET", "/users/abc", token, nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGet_401_NoToken(t *testing.T) {
	mux, _, _ := newTestMux()

	rr := doJSON(mux, "GET", "/users/1", "", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUpdate_200(t *testing.T) {
	mux, store, token := newTestMux()

	u, _ := store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "PUT", "/users/"+itoa(u.ID), token,
		models.User{Name: "Bob", Email: "b@x.com", Age: 30})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}

	var got models.User
	_ = json.Unmarshal(rr.Body.Bytes(), &got)
	if got.Name != "Bob" || got.ID != u.ID {
		t.Errorf("unexpected body: %+v", got)
	}
}

func TestUpdate_404(t *testing.T) {
	mux, _, token := newTestMux()

	rr := doJSON(mux, "PUT", "/users/999", token,
		models.User{Name: "Bob", Email: "b@x.com", Age: 30})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdate_400_Validation(t *testing.T) {
	mux, store, token := newTestMux()

	u, _ := store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "PUT", "/users/"+itoa(u.ID), token,
		models.User{Name: "Bob", Email: "bad", Age: 30})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdate_401_NoToken(t *testing.T) {
	mux, store, _ := newTestMux()

	u, _ := store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "PUT", "/users/"+itoa(u.ID), "",
		models.User{Name: "Bob", Email: "b@x.com", Age: 30})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestDelete_204(t *testing.T) {
	mux, store, token := newTestMux()

	u, _ := store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "DELETE", "/users/"+itoa(u.ID), token, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	rr = doJSON(mux, "GET", "/users/"+itoa(u.ID), token, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rr.Code)
	}
}

func TestDelete_404(t *testing.T) {
	mux, _, token := newTestMux()

	rr := doJSON(mux, "DELETE", "/users/999", token, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDelete_401_NoToken(t *testing.T) {
	mux, store, _ := newTestMux()

	u, _ := store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "DELETE", "/users/"+itoa(u.ID), "", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestList_200(t *testing.T) {
	mux, _, token := newTestMux()

	rr := doJSON(mux, "GET", "/users", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
}

func TestList_401_NoToken(t *testing.T) {
	mux, _, _ := newTestMux()

	rr := doJSON(mux, "GET", "/users", "", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestList_FilterAndPaginate(t *testing.T) {
	mux, store, token := newTestMux()

	for _, u := range []models.User{
		{Name: "Alice", Email: "a@x.com", Age: 20},
		{Name: "Bob", Email: "b@x.com", Age: 30},
		{Name: "Alicia", Email: "c@x.com", Age: 40},
	} {
		_, _ = store.Create(u)
	}

	rr := doJSON(mux, "GET", "/users?limit=2&offset=1", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var items []models.User
	_ = json.Unmarshal(rr.Body.Bytes(), &items)
	if len(items) != 2 {
		t.Errorf("pagination: expected 2 items, got %d", len(items))
	}

	rr = doJSON(mux, "GET", "/users?name=ali", token, nil)
	_ = json.Unmarshal(rr.Body.Bytes(), &items)
	if len(items) != 2 {
		t.Errorf("name filter: expected 2 items, got %d", len(items))
	}

	rr = doJSON(mux, "GET", "/users?min_age=25&max_age=35", token, nil)
	_ = json.Unmarshal(rr.Body.Bytes(), &items)
	if len(items) != 1 || items[0].Name != "Bob" {
		t.Errorf("age filter: got %+v", items)
	}

	rr = doJSON(mux, "GET", "/users?limit=9999", token, nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	rr = doJSON(mux, "GET", "/users?min_age=50&max_age=10", token, nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for min>max, got %d", rr.Code)
	}

	rr = doJSON(mux, "GET", "/users?offset=-5", token, nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative offset, got %d", rr.Code)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
