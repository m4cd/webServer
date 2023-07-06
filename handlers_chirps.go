package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/m4cd/webServer/internal/webapi"
)

func ValidateChirpHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := chirpBodyParameters{}
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

func PostChirpsHandler(apicfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		db, _ := webapi.NewDB(databasePath)
		decoder := json.NewDecoder(r.Body)
		params := chirpBodyParameters{}
		err := decoder.Decode(&params)

		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		//Authentication
		header := r.Header.Get("Authorization")
		if header == "" {
			fmt.Println("No \"Authorization\" header present")
			respondWithCode(w, 401)
			return
		}
		requestTokenString := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.ParseWithClaims(requestTokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(apicfg.jwtSecret), nil })

		if err != nil {
			fmt.Println("Invalid token:", err)
			respondWithCode(w, 401)
			return
		}

		issuer, err := token.Claims.GetIssuer()
		if err != nil {
			fmt.Println("Error extracting issuer:", err)
			respondWithCode(w, 401)
			return
		}
		if apicfg.accessTokenIssuer != issuer {
			fmt.Println("No access token in the header. Request rejected.")
			respondWithCode(w, 401)
			return
		}

		authorIDstr, err := token.Claims.GetSubject()
		//fmt.Println(authorIDstr)
		if err != nil {
			fmt.Println("Error while extracting author ID")
			respondWithCode(w, 401)
			return
		}
		authorID, _ := strconv.Atoi(authorIDstr)
		//fmt.Println(authorID)
		newChirp, err := db.CreateChirp(params.BodyJSON, authorID)

		if err != nil {
			respondWithError(w, 500, "Something went wrong while creating chirp")
			return
		}
		//fmt.Println(newChirp)
		respondWithJSON(w, 201, newChirp)
	}
}

func GetChirpsHandler(w http.ResponseWriter, r *http.Request) {

	db, _ := webapi.NewDB(databasePath)
	dbChirps, err := db.GetChirps()

	if err != nil {
		respondWithError(w, 500, "Something went wrong while getting chirps")
		return
	}

	authorIDstr := r.URL.Query().Get("author_id")
	authorID, _ := strconv.Atoi(authorIDstr)
	sortParam := r.URL.Query().Get("sort")

	ascending := true

	if sortParam == "desc" {
		ascending = false
	} else if sortParam != "asc" && sortParam != "" {
		fmt.Printf("Incorrect \"sort\" parameter value: %v\n", sortParam)
	}

	authorsChirps := []webapi.Chirp{}

	if authorID != 0 {
		for _, chirp := range dbChirps {
			if chirp.AuthorID == authorID {
				authorsChirps = append(authorsChirps, chirp)
			}
		}
		/*
			if !ascending {
				sort.Slice(authorsChirps[:], func(i, j int) bool {
					return authorsChirps[i].ID > authorsChirps[j].ID
				})
			}*/

		respondWithJSON(w, 200, authorsChirps)
		return
	}

	if !ascending {
		sort.Slice(dbChirps[:], func(i, j int) bool {
			return dbChirps[i].ID > dbChirps[j].ID
		})
	}

	respondWithJSON(w, 200, dbChirps)

}

func GetChirpByIdHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := webapi.NewDB(databasePath)

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
	respondWithJSON(w, 200, chirp)
}

func DeleteChirpByIdHandler(apicfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestTokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		token, err := jwt.ParseWithClaims(requestTokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(apicfg.jwtSecret), nil })

		if err != nil {
			fmt.Println("Invalid token:", err)
			respondWithCode(w, 403)
			return
		}

		issuer, err := token.Claims.GetIssuer()
		if err != nil {
			fmt.Println("Error extracting issuer:", err)
			respondWithCode(w, 403)
			return
		}
		if apicfg.refreshTokenIssuer == issuer {
			fmt.Println("Refresh token in the header. Request rejected.")
			respondWithCode(w, 403)
			return
		}
		expirationTime, _ := token.Claims.GetExpirationTime()

		if expirationTime.Compare(time.Now()) != 1 {
			fmt.Println("Expired token:", err)
			respondWithCode(w, 403)
			return
		}

		idString, _ := token.Claims.GetSubject()
		id, _ := strconv.Atoi(idString)

		db, _ := webapi.NewDB(databasePath)

		dbChirps, _ := db.GetChirps()
		chirpIDstr := chi.URLParam(r, "chirpID")
		chirpIDint, _ := strconv.Atoi(chirpIDstr)

		for _, chirp := range dbChirps {
			if chirp.AuthorID == id && chirp.ID == chirpIDint {
				if db.DeleteChirp(chirpIDint) == nil {
					respondWithJSON(w, 200, chirp)
					return
				} else {
					break
				}

			}
		}

		respondWithCode(w, 403)
	}
}
