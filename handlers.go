package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/m4cd/webServer/internal/database"
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

func customValidateChirpHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := bodyParameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	const maxLength = 140
	if len(params.BodyJSON) > maxLength {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	bodyWords := strings.Split(params.BodyJSON, " ")
	bodyWordsLower := strings.ToLower(params.BodyJSON)
	bodyWordsArray := strings.Split(bodyWordsLower, " ")

	profane := [3]string{"kerfuffle", "sharbert", "fornax"}

	for i, word := range bodyWordsArray {
		for _, prof := range profane {
			if word == prof {
				bodyWords[i] = "****"
			}
		}
	}

	cleanBody := strings.Join(bodyWords, " ")
	respondWithJSON(w, 200, cleanedBody{CleanedBody: cleanBody})
}

func customPostChirpsHandler(w http.ResponseWriter, r *http.Request) {

	db, _ := database.NewDB(databasePath)
	decoder := json.NewDecoder(r.Body)
	params := bodyParameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	newChirp, err := db.CreateChirp(params.BodyJSON)
	if err != nil {
		respondWithError(w, 500, "Something went wrong while creating chirp")
		return
	}

	respondWithJSON(w, 201, newChirp)
}

func customGetChirpsHandler(w http.ResponseWriter, r *http.Request) {

	db, _ := database.NewDB(databasePath)

	dbChirps, err := db.GetChirps()

	if err != nil {
		respondWithError(w, 500, "Something went wrong while getting chirps")
		return
	}
	respondWithJSON(w, 200, dbChirps)

}

func customGetChirpByIdHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.NewDB(databasePath)

	dbChirps, err := db.GetChirps()

	if err != nil {
		respondWithError(w, 500, "Something went wrong while getting a chirp by ID")
		return
	}
	chirpIDstr := chi.URLParam(r, "chirpID")
	chirpIDint, err := strconv.Atoi(chirpIDstr)

	if err != nil {
		fmt.Println("Error while converting chirp ID from string to int")
	}

	if len(dbChirps) < chirpIDint {
		w.WriteHeader(404)
		return
	}

	chirp := dbChirps[chirpIDint-1]
	fmt.Println(chirp)
	respondWithJSON(w, 200, chirp)
}

func customPostCreateUser(w http.ResponseWriter, r *http.Request) {
	db, _ := database.NewDB(databasePath)
	decoder := json.NewDecoder(r.Body)
	params := bodyUser{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	newUser, err := db.CreateUser(params.Email)
	if err != nil {
		respondWithError(w, 500, "Something went wrong while creating a user")
		return
	}

	respondWithJSON(w, 201, newUser)
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
