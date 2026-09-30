package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/data", dataHandler)
	mux.HandleFunc("/", rootHandler)

	handler := loggingMiddleware(mux)

	fmt.Println("HTTP сервер с middleware слушает порт 8080")
	http.ListenAndServe(":8080", handler)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Привет из /hello")
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Это /data")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Главная страница")
}
