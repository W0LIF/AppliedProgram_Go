package main

import (
	"log"
	"net/http"
	"os"

	"laba9/handlers"
	"laba9/storage"
)

func main() {
	store, err := buildStorage()
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}

	tokens := storage.NewTokenStore()
	userH := handlers.NewUserHandler(store)
	authH := handlers.NewAuthHandler(store, tokens)

	mux := http.NewServeMux()

	// Публичные:
	mux.HandleFunc("POST /auth/register", handlers.Wrap(authH.Register))
	mux.HandleFunc("POST /auth/login", handlers.Wrap(authH.Login))
	mux.HandleFunc("POST /auth/logout", handlers.Wrap(authH.Logout))

	// Защищённые:
	mux.HandleFunc("GET /users", handlers.Wrap(handlers.RequireAuth(tokens, userH.List)))
	mux.HandleFunc("GET /users/{id}", handlers.Wrap(handlers.RequireAuth(tokens, userH.Get)))
	mux.HandleFunc("POST /users", handlers.Wrap(handlers.RequireAuth(tokens, userH.Create)))
	mux.HandleFunc("PUT /users/{id}", handlers.Wrap(handlers.RequireAuth(tokens, userH.Update)))
	mux.HandleFunc("DELETE /users/{id}", handlers.Wrap(handlers.RequireAuth(tokens, userH.Delete)))

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func buildStorage() (storage.UserStorage, error) {
	switch os.Getenv("STORAGE") {
	case "postgres":
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			dsn = "postgres://lab:lab@localhost:5434/lab8?sslmode=disable"
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
