package handlers

import (
	"errors"
	"log"
	"net/http"

	"laba10/models"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Wrap(next HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			handleError(w, err)
		}
	}
}

func handleError(w http.ResponseWriter, err error) {
	var ve *models.ValidationError
	if errors.As(err, &ve) {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":   "validation failed",
			"details": ve.Fields,
		})
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.cause != nil {
			log.Printf("app error [%d]: %s | cause: %v", appErr.Status, appErr.Message, appErr.cause)
		} else {
			log.Printf("app error [%d]: %s", appErr.Status, appErr.Message)
		}
		body := map[string]any{"error": appErr.Message}
		if appErr.Details != nil {
			body["details"] = appErr.Details
		}
		writeJSON(w, appErr.Status, body)
		return
	}

	log.Printf("unhandled error: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = writeJSONEncode(w, data)
}
