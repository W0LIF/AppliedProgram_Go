package handlers

import (
	"context"
	"net/http"
	"strings"

	"laba10/auth"
)

type ctxKey string

const (
	ctxUserID ctxKey = "uid"
	ctxRole   ctxKey = "role"
)

const (
	csrfCookieName    = "csrf_token"
	csrfHeaderName    = "X-CSRF-Token"
	sessionCookieName = "session_id"
)

func RequireAuth(tm *auth.TokenManager, ss *auth.SessionStore, next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if token := extractBearer(r); token != "" {
			claims, err := tm.Parse(token)
			if err != nil {
				return Unauthorized("invalid or expired token")
			}
			ctx := withAuth(r.Context(), claims.UserID, claims.Role)
			return next(w, r.WithContext(ctx))
		}

		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			return Unauthorized("missing Authorization header or session cookie")
		}
		sess, ok := ss.Get(cookie.Value)
		if !ok {
			return Unauthorized("invalid session")
		}

		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if err := checkCSRF(r); err != nil {
				return err
			}
		}

		ctx := withAuth(r.Context(), sess.UserID, sess.Role)
		return next(w, r.WithContext(ctx))
	}
}

func RequireRole(next HandlerFunc, roles ...string) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		role, _ := r.Context().Value(ctxRole).(string)
		for _, allowed := range roles {
			if role == allowed {
				return next(w, r)
			}
		}
		return Forbidden("forbidden")
	}
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func withAuth(ctx context.Context, userID int, role string) context.Context {
	ctx = context.WithValue(ctx, ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxRole, role)
	return ctx
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(ctxUserID).(int)
	return id, ok
}
