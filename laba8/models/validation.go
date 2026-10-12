package models

import (
	"fmt"
	"net/mail"
	"strings"
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %d field(s)", len(e.Fields))
}

func (u *User) Validate() error {
	fields := make(map[string]string)

	name := strings.TrimSpace(u.Name)
	switch {
	case name == "":
		fields["name"] = "name is required"
	case len([]rune(name)) < 2:
		fields["name"] = "name must be at least 2 characters"
	case len([]rune(name)) > 100:
		fields["name"] = "name must be at most 100 characters"
	}
	u.Name = name

	email := strings.TrimSpace(u.Email)
	switch {
	case email == "":
		fields["email"] = "email is required"
	default:
		if _, err := mail.ParseAddress(email); err != nil {
			fields["email"] = "email is invalid"
		}
	}
	u.Email = email

	if u.Age < 1 || u.Age > 150 {
		fields["age"] = "age must be between 1 and 150"
	}

	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}
