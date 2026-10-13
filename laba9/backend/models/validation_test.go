package models

import "testing"

func TestUserValidate_OK(t *testing.T) {
	u := User{Name: "  Alice  ", Email: "a@x.com", Age: 25}
	if err := u.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if u.Name != "Alice" {
		t.Errorf("name should be trimmed, got %q", u.Name)
	}
}

func TestUserValidate_Errors(t *testing.T) {
	cases := []struct {
		name  string
		user  User
		field string
	}{
		{"empty name", User{Name: "", Email: "a@x.com", Age: 20}, "name"},
		{"short name", User{Name: "A", Email: "a@x.com", Age: 20}, "name"},
		{"bad email", User{Name: "Bob", Email: "not-email", Age: 20}, "email"},
		{"age too low", User{Name: "Bob", Email: "a@x.com", Age: 0}, "age"},
		{"age too high", User{Name: "Bob", Email: "a@x.com", Age: 200}, "age"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.user.Validate()
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
			if _, ok := ve.Fields[c.field]; !ok {
				t.Errorf("expected error on field %q, got %v", c.field, ve.Fields)
			}
		})
	}
}

func TestUserValidate_MultipleErrors(t *testing.T) {
	u := User{Name: "", Email: "broken", Age: -1}
	err := u.Validate()
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d: %v", len(ve.Fields), ve.Fields)
	}
}
