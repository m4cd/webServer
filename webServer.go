package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	godotenv.Load()

	if *dbg {
		os.Remove(databasePath)
	}

	serverRoot := "."
	serverPort := "8080"

	apiCfg := apiConfig{
		fileserverHits:         0,
		jwtSecret:              os.Getenv("JWT_SECRET"),
		accessTokenExpiration:  3600, //1 hour
		accessTokenIssuer:      "chirpy-access",
		refreshTokenExpiration: 60 * 24 * 3600, //60 days
		refreshTokenIssuer:     "chirpy-refresh",
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
	routerAPI.Post("/chirps", PostChirpsHandler(&apiCfg))
	routerAPI.Get("/chirps", GetChirpsHandler)
	routerAPI.Get("/chirps", GetChirpsHandler)
	routerAPI.Get("/chirps/{chirpID}", GetChirpByIdHandler)
	routerAPI.Delete("/chirps/{chirpID}", DeleteChirpByIdHandler(&apiCfg))
	routerAPI.Post("/users", PostCreateUser)
	routerAPI.Post("/login", UserLogin(&apiCfg))
	routerAPI.Put("/users", PutUpdateUser(&apiCfg))
	routerAPI.Post("/refresh", PostRefreshToken(&apiCfg))
	routerAPI.Post("/revoke", PostRevokeToken(&apiCfg))

	router.Mount("/admin", routerAdmin)
	router.Mount("/api", routerAPI)

	corsMux := middlewareCors(router)

	server := http.Server{
		Handler: corsMux,
		Addr:    ":" + serverPort,
	}

	server.ListenAndServe()
}
