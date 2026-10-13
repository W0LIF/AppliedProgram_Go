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

func (r *RegisterRequest) Validate() error {
	u := User{Name: r.Name, Email: r.Email, Age: r.Age}
	fields := make(map[string]string)
	if err := u.Validate(); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			for k, v := range ve.Fields {
				fields[k] = v
			}
		}
	}
	r.Name, r.Email, r.Age = u.Name, u.Email, u.Age
	if len(r.Password) < 6 {
		fields["password"] = "password must be at least 6 characters"
	}
	if len(r.Password) > 50 {
		fields["password"] = "password must be at most 50 characters"
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}
