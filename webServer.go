package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	serverRoot := "."
	serverPort := "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	router := chi.NewRouter()
	routerAPI := chi.NewRouter()

	/*
		mux := http.NewServeMux()
		mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(serverRoot))))
		mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(serverRoot)))))
		mux.HandleFunc("/healthz", customHandler)
		mux.HandleFunc("/metrics", apiCfg.customMetricsHandler)
		corsMux := middlewareCors(mux)
	*/

	router.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(serverRoot)))))
	router.Handle("/app", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(serverRoot)))))
	routerAPI.Get("/healthz", customHandler)
	routerAPI.Get("/metrics", apiCfg.customMetricsHandler)

	router.Mount("/api", routerAPI)

	corsMux := middlewareCors(router)

	server := http.Server{
		Handler: corsMux,
		Addr:    ":" + serverPort,
	}

	server.ListenAndServe()
}
