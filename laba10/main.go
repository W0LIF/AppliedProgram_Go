package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"

	"laba10/auth"
	"laba10/handlers"
	"laba10/storage"
)

func main() {
	store, err := buildStorage()
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}

	tokens := auth.NewTokenManager(os.Getenv("JWT_SECRET"))
	sessions := auth.NewSessionStore()

	userH := handlers.NewUserHandler(store)
	authH := handlers.NewAuthHandler(store, tokens, sessions)

	authMw := func(next handlers.HandlerFunc) handlers.HandlerFunc {
		return handlers.RequireAuth(tokens, sessions, next)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/register", handlers.Wrap(authH.Register))
	mux.HandleFunc("POST /auth/login", handlers.Wrap(authH.Login))
	mux.HandleFunc("POST /auth/session-login", handlers.Wrap(authH.SessionLogin))
	mux.HandleFunc("POST /auth/session-logout", handlers.Wrap(authH.SessionLogout))

	mux.HandleFunc("GET /users", handlers.Wrap(
		handlers.EnsureCSRFCookie(
			authMw(userH.List))))

	mux.HandleFunc("GET /users/{id}", handlers.Wrap(
		handlers.EnsureCSRFCookie(
			authMw(userH.Get))))

	mux.HandleFunc("POST /users", handlers.Wrap(
		handlers.EnsureCSRFCookie(
			authMw(handlers.RequireRole(userH.Create, "admin")))))

	mux.HandleFunc("PUT /users/{id}", handlers.Wrap(
		handlers.EnsureCSRFCookie(
			authMw(handlers.RequireRole(userH.Update, "admin")))))

	mux.HandleFunc("DELETE /users/{id}", handlers.Wrap(
		handlers.EnsureCSRFCookie(
			authMw(handlers.RequireRole(userH.Delete, "admin")))))

	caPEM, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("read CA: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		log.Fatal("bad CA cert")
	}

	tlsCfg := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  pool,
		MinVersion: tls.VersionTLS12,
	}

	srv := &http.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: tlsCfg,
	}

	log.Println("HTTPS listening on :8443 (mTLS required)")
	if err := srv.ListenAndServeTLS("certs/server.crt", "certs/server.key"); err != nil {
		log.Fatal(err)
	}
}

func buildStorage() (storage.UserStorage, error) {
	switch os.Getenv("STORAGE") {
	case "postgres":
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			dsn = "postgres://lab:lab@localhost:5434/lab10?sslmode=disable"
		}
		pg, err := storage.NewPostgresStorage(dsn)
		if err != nil {
			return nil, err
		}
		if err := pg.Migrate(); err != nil {
			return nil, err
		}
		log.Println("storage: postgres")
		return pg, nil
	default:
		log.Println("storage: memory")
		return storage.NewMemoryStorage(), nil
	}
}
