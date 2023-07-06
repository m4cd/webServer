package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/m4cd/webServer/internal/webapi"
)

func PolkaChirpyRedWebhook(apicfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			fmt.Println("No \"Authorization\" header present")
			respondWithCode(w, 401)
			return
		}

		requestApiKeyString := strings.TrimPrefix(header, "ApiKey ")

		if requestApiKeyString != apicfg.polkaApiKey {
			respondWithCode(w, 401)
			return
		}

		decoder := json.NewDecoder(r.Body)
		params := chipryRedWebhookBody{}
		err := decoder.Decode(&params)

		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		if params.Event != "user.upgraded" {
			//respondWithCode(w, 200)
			respondWithJSON(w, 200, webapi.EmptyStruct{})
			return
		}

		db, _ := webapi.NewDB(databasePath)
		err = db.ChirpyRedUpdateUser(params.Data.UserID)

		if err != nil {
			fmt.Println("Error while updating Chirpy Red value:", err)
			respondWithCode(w, 404)
			return
		}

		//respondWithCode(w, 200)
		respondWithJSON(w, 200, webapi.EmptyStruct{})
	}
}
