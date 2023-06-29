package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/m4cd/webServer/internal/webapi"
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

func PostChirpsHandler(w http.ResponseWriter, r *http.Request) {

	db, _ := webapi.NewDB(databasePath)
	decoder := json.NewDecoder(r.Body)
	params := chirpBodyParameters{}
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

func GetChirpsHandler(w http.ResponseWriter, r *http.Request) {

	db, _ := webapi.NewDB(databasePath)

	dbChirps, err := db.GetChirps()

	if err != nil {
		respondWithError(w, 500, "Something went wrong while getting chirps")
		return
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
	fmt.Println(chirp)
	respondWithJSON(w, 200, chirp)
}

func PostCreateUser(w http.ResponseWriter, r *http.Request) {
	db, _ := webapi.NewDB(databasePath)
	decoder := json.NewDecoder(r.Body)
	params := PostUser{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	newUser, err := db.CreateUser(params.Password, params.Email)
	if err != nil {
		respondWithError(w, 500, "Something went wrong while creating a user")
		return
	}

	respondWithJSON(w, 201, newUser)
}

func UserLogin(apicfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, _ := webapi.NewDB(databasePath)
		decoder := json.NewDecoder(r.Body)
		params := PostUser{}
		err := decoder.Decode(&params)

		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		UserLoggedIn, err := db.VerifyCredentials(params.Email, params.Password)
		if err == nil {
			currentTime := time.Now()
			subject := strconv.Itoa(UserLoggedIn.ID)

			UserLoggedIn.Token = generateToken(
				jwt.NewNumericDate(currentTime),
				apicfg.accessTokenIssuer,
				jwt.NewNumericDate(currentTime.Add(time.Duration(apicfg.accessTokenExpiration)*time.Second)),
				subject,
				apicfg)
			UserLoggedIn.RefreshToken = generateToken(
				jwt.NewNumericDate(currentTime),
				apicfg.refreshTokenIssuer,
				jwt.NewNumericDate(currentTime.Add(time.Duration(apicfg.refreshTokenExpiration)*time.Second)),
				subject,
				apicfg)
			respondWithJSON(w, 200, UserLoggedIn)
			return
		}

		respondWithError(w, 401, "Unauthorized")
	}
}

func generateToken(currentTime *jwt.NumericDate, issuer string, expirationTime *jwt.NumericDate, subject string, apicfg *apiConfig) string {
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  currentTime,
		ExpiresAt: expirationTime,
		Subject:   subject,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(apicfg.jwtSecret))

	if err != nil {
		fmt.Println("Error while token signing:", err)
	}
	return signedToken
}

func PutUpdateUser(apicfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, _ := webapi.NewDB(databasePath)
		decoder := json.NewDecoder(r.Body)
		params := PostUser{}
		err := decoder.Decode(&params)

		fmt.Printf("Body: %v\n", params)

		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		requestTokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		//fmt.Printf("requestTokenString: %v\n", requestTokenString)

		token, err := jwt.ParseWithClaims(requestTokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(apicfg.jwtSecret), nil })

		if err != nil {
			fmt.Println("Invalid token:", err)
			//respondWithError(w, 401, "Unauthorized")
			respondWithCode(w, 401)
			return
		}

		expirationTime, _ := token.Claims.GetExpirationTime()
		idString, _ := token.Claims.GetSubject()
		id, _ := strconv.Atoi(idString)

		if expirationTime.Compare(time.Now()) != 1 {
			fmt.Println("Expired token:", err)
			//respondWithError(w, 401, "Unauthorized")
			respondWithCode(w, 401)
			return
		}

		UserModified, err := db.UpdateUser(id, params.Email, params.Password)
		if err != nil {
			fmt.Println("Error while user update:", err)
			//respondWithError(w, 401, "Unauthorized")
			respondWithCode(w, 401)
			return
		}
		respondWithJSON(w, 200, UserModified)
	}
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
