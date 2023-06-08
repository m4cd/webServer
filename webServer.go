package main

import (
	"net/http"
)

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	serverRoot := "."
	serverPort := "8080"

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir(serverRoot)))

	corsMux := middlewareCors(mux)

	server := http.Server{
		Handler: corsMux,
		Addr:    ":" + serverPort,
	}

	server.ListenAndServe()
}
