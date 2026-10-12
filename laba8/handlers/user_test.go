package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"laba8/models"
	"laba8/storage"
)

func newTestMux() (*http.ServeMux, *storage.MemoryStorage) {
	store := storage.NewMemoryStorage()
	h := NewUserHandler(store)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", Wrap(h.List))
	mux.HandleFunc("GET /users/{id}", Wrap(h.Get))
	mux.HandleFunc("POST /users", Wrap(h.Create))
	mux.HandleFunc("PUT /users/{id}", Wrap(h.Update))
	mux.HandleFunc("DELETE /users/{id}", Wrap(h.Delete))
	return mux, store
}

func doJSON(mux *http.ServeMux, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestCreate_201(t *testing.T) {
	mux, _ := newTestMux()
	rr := doJSON(mux, "POST", "/users", models.User{Name: "Alice", Email: "a@x.com", Age: 25})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body)
	}
	var got models.User
	_ = json.Unmarshal(rr.Body.Bytes(), &got)
	if got.ID != 1 || got.Name != "Alice" {
		t.Errorf("unexpected body: %+v", got)
	}
}

func TestCreate_400_InvalidJSON(t *testing.T) {
	mux, _ := newTestMux()
	req := httptest.NewRequest("POST", "/users", strings.NewReader("not-json"))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreate_400_Validation(t *testing.T) {
	mux, _ := newTestMux()
	rr := doJSON(mux, "POST", "/users", models.User{Name: "", Email: "x", Age: -1})
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

func TestGet_200(t *testing.T) {
	mux, store := newTestMux()
	_, _ = store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "GET", "/users/1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGet_404(t *testing.T) {
	mux, _ := newTestMux()
	rr := doJSON(mux, "GET", "/users/999", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGet_400_InvalidID(t *testing.T) {
	mux, _ := newTestMux()
	rr := doJSON(mux, "GET", "/users/abc", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdate_200(t *testing.T) {
	mux, store := newTestMux()
	_, _ = store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "PUT", "/users/1", models.User{Name: "Bob", Email: "b@x.com", Age: 30})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
}

func TestUpdate_404(t *testing.T) {
	mux, _ := newTestMux()
	rr := doJSON(mux, "PUT", "/users/999", models.User{Name: "Bob", Email: "b@x.com", Age: 30})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdate_400_Validation(t *testing.T) {
	mux, store := newTestMux()
	_, _ = store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "PUT", "/users/1", models.User{Name: "Bob", Email: "bad", Age: 30})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDelete_204(t *testing.T) {
	mux, store := newTestMux()
	_, _ = store.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})

	rr := doJSON(mux, "DELETE", "/users/1", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	rr = doJSON(mux, "GET", "/users/1", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rr.Code)
	}
}

func TestDelete_404(t *testing.T) {
	mux, _ := newTestMux()
	rr := doJSON(mux, "DELETE", "/users/999", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestList_FilterAndPaginate(t *testing.T) {
	mux, store := newTestMux()
	for _, u := range []models.User{
		{Name: "Alice", Email: "a@x.com", Age: 20},
		{Name: "Bob", Email: "b@x.com", Age: 30},
		{Name: "Alicia", Email: "c@x.com", Age: 40},
	} {
		_, _ = store.Create(u)
	}

	rr := doJSON(mux, "GET", "/users?limit=2&offset=1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var items []models.User
	_ = json.Unmarshal(rr.Body.Bytes(), &items)
	if len(items) != 2 || items[0].Name != "Bob" {
		t.Errorf("pagination body: %+v", items)
	}

	rr = doJSON(mux, "GET", "/users?name=ali", nil)
	_ = json.Unmarshal(rr.Body.Bytes(), &items)
	if len(items) != 2 {
		t.Errorf("name filter: got %d items", len(items))
	}

	rr = doJSON(mux, "GET", "/users?limit=9999", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
