package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func customHealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) customMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	hits := fmt.Sprintf(`<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>`, cfg.fileserverHits)
	w.Write([]byte(hits))

}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func respondWithError(w http.ResponseWriter, code int, errorMessage string) {
	respondWithJSON(w, code, errorParameters{Error: errorMessage})
}

func respondWithJSON(w http.ResponseWriter, code int, jsonStruct interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	respData, _ := json.Marshal(jsonStruct)
	w.Write(respData)
}

func respondWithCode(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
}
