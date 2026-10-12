package main

import (
	"log"
	"net/http"
	"os"

	"laba8/handlers"
	"laba8/storage"
)

func main() {
	store, err := buildStorage()
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}
	h := handlers.NewUserHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", handlers.Wrap(h.List))
	mux.HandleFunc("GET /users/{id}", handlers.Wrap(h.Get))
	mux.HandleFunc("POST /users", handlers.Wrap(h.Create))
	mux.HandleFunc("PUT /users/{id}", handlers.Wrap(h.Update))
	mux.HandleFunc("DELETE /users/{id}", handlers.Wrap(h.Delete))

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
