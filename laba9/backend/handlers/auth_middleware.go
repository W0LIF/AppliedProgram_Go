package handlers

import (
	"context"
	"net/http"
	"strings"

	"laba9/storage"
)

type ctxKey string

const ctxUserID ctxKey = "userID"

func RequireAuth(tokens *storage.TokenStore, next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		token := extractToken(r)
		if token == "" {
			return Unauthorized("missing Authorization header")
		}
		userID, ok := tokens.Get(token)
		if !ok {
			return Unauthorized("invalid or expired token")
		}
		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		return next(w, r.WithContext(ctx))
	}
}

func extractToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(ctxUserID).(int)
	return id, ok
}
