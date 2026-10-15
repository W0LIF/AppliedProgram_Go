package handlers

import (
	"errors"
	"net/http"

	"laba10/storage"
)

type AppError struct {
	Status  int
	Message string
	Details any
	cause   error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.cause }

func BadRequest(msg string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Status: http.StatusForbidden, Message: msg}
}

func NotFound(msg string) *AppError {
	return &AppError{Status: http.StatusNotFound, Message: msg}
}

func Internal(cause error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Message: "internal error",
		cause:   cause,
	}
}

func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

func mapStorageError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, storage.ErrNotFound) {
		return NotFound("user not found")
	}
	return Internal(err)
}
