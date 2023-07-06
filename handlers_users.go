package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/m4cd/webServer/internal/webapi"
)

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

		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		requestTokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

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
		if apicfg.refreshTokenIssuer == issuer {
			fmt.Println("Refresh token in the header. Request rejected.")
			respondWithCode(w, 401)
			return
		}
		expirationTime, _ := token.Claims.GetExpirationTime()
		idString, _ := token.Claims.GetSubject()
		id, _ := strconv.Atoi(idString)

		if expirationTime.Compare(time.Now()) != 1 {
			fmt.Println("Expired token:", err)
			respondWithCode(w, 401)
			return
		}

		UserModified, err := db.UpdateUser(id, params.Email, params.Password)
		if err != nil {
			fmt.Println("Error while user update:", err)
			respondWithCode(w, 401)
			return
		}
		respondWithJSON(w, 200, UserModified)
	}
}

func PostRefreshToken(apicfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		if apicfg.refreshTokenIssuer != issuer {
			fmt.Println("No refresh token in the header. Request rejected.")
			respondWithCode(w, 401)
			return
		}

		db, _ := webapi.NewDB(databasePath)
		revokedTokenFound := db.CheckToken(requestTokenString)
		if revokedTokenFound == true {
			fmt.Println("Token revoked.")
			respondWithCode(w, 401)
			return
		}

		currentTime := time.Now()
		subject, _ := token.Claims.GetSubject()
		accessToken := generateToken(
			jwt.NewNumericDate(currentTime),
			apicfg.accessTokenIssuer,
			jwt.NewNumericDate(currentTime.Add(time.Duration(apicfg.accessTokenExpiration)*time.Second)),
			subject,
			apicfg)

		respondWithJSON(w, 200, Token{Token: accessToken})
	}
}

func PostRevokeToken(apicfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
		if apicfg.refreshTokenIssuer != issuer {
			fmt.Println("No refresh token in the header. Request rejected.")
			respondWithCode(w, 401)
			return
		}
		db, _ := webapi.NewDB(databasePath)
		db.RevokeToken(requestTokenString, *jwt.NewNumericDate(time.Now()))
		//respondWithCode(w, 200)
		respondWithJSON(w, 200, Token{Token: requestTokenString})
	}
}
