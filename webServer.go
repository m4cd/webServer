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
	routerAdmin := chi.NewRouter()
	routerAPI := chi.NewRouter()

	/*
		mux := http.NewServeMux()
		mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(serverRoot))))
		mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(serverRoot)))))
		mux.HandleFunc("/healthz", customHandler)
		mux.HandleFunc("/metrics", apiCfg.customMetricsHandler)
		corsMux := middlewareCors(mux)
	*/

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(serverRoot))))
	router.Handle("/app/*", fsHandler)
	router.Handle("/app", fsHandler)

	//router.Handle("/*", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(serverRoot))))
	//router.Handle("/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(serverRoot))))

	routerAPI.Get("/healthz", customHealthzHandler)
	routerAdmin.Get("/metrics", apiCfg.customMetricsHandler)
	//routerAPI.Post("/validate_chirp", customValidateChirpHandler)
	routerAPI.Post("/chirps", customPostChirpsHandler)
	routerAPI.Get("/chirps", customGetChirpsHandler)
	routerAPI.Get("/chirps", customGetChirpsHandler)
	routerAPI.Get("/chirps/{chirpID}", customGetChirpByIdHandler)
	routerAPI.Post("/users", customPostCreateUser)

	router.Mount("/admin", routerAdmin)
	router.Mount("/api", routerAPI)

	corsMux := middlewareCors(router)

	server := http.Server{
		Handler: corsMux,
		Addr:    ":" + serverPort,
	}

	server.ListenAndServe()
}
