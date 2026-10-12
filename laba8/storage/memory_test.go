package storage

import (
	"errors"
	"testing"

	"laba8/models"
)

func TestMemoryStorage_CRUD(t *testing.T) {
	s := NewMemoryStorage()

	u, err := s.Create(models.User{Name: "Alice", Email: "a@x.com", Age: 25})
	if err != nil || u.ID != 1 {
		t.Fatalf("create: got %+v, err %v", u, err)
	}

	got, err := s.Get(1)
	if err != nil || got.Name != "Alice" {
		t.Fatalf("get: got %+v, err %v", got, err)
	}

	updated, err := s.Update(1, models.User{Name: "Alice2", Email: "a2@x.com", Age: 26})
	if err != nil || updated.Name != "Alice2" {
		t.Fatalf("update: got %+v, err %v", updated, err)
	}

	if err := s.Delete(1); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.Get(1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryStorage_List_FilterAndPaginate(t *testing.T) {
	s := NewMemoryStorage()
	for _, u := range []models.User{
		{Name: "Alice", Email: "a@x.com", Age: 20},
		{Name: "Bob", Email: "b@x.com", Age: 30},
		{Name: "Alicia", Email: "c@x.com", Age: 40},
	} {
		_, _ = s.Create(u)
	}

	got, _ := s.List(UserFilter{Name: "ali"})
	if len(got) != 2 {
		t.Errorf("name filter: expected 2, got %d", len(got))
	}

	got, _ = s.List(UserFilter{MinAge: 25, MaxAge: 35})
	if len(got) != 1 || got[0].Name != "Bob" {
		t.Errorf("age filter: got %+v", got)
	}

	got, _ = s.List(UserFilter{Limit: 2, Offset: 1})
	if len(got) != 2 || got[0].Name != "Bob" {
		t.Errorf("pagination: got %+v", got)
	}

	got, _ = s.List(UserFilter{Limit: 5, Offset: 100})
	if len(got) != 0 {
		t.Errorf("offset out of range: got %d items", len(got))
	}
}
