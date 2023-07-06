package main

type apiConfig struct {
	fileserverHits         int
	jwtSecret              string
	accessTokenExpiration  int
	accessTokenIssuer      string
	refreshTokenExpiration int
	refreshTokenIssuer     string
	polkaApiKey            string
}

type errorParameters struct {
	Error string `json:"error"`
}

type chirpBodyParameters struct {
	BodyJSON string `json:"body"`
}

type cleanedBody struct {
	CleanedBody string `json:"cleaned_body"`
}

type PostUser struct {
	Password   string `json:"password"`
	Email      string `json:"email"`
	Expiration int    `json:"expires_in_seconds,omitempty"`
	ChirpyRed  bool   `json:"is_chirpy_red,omitempty"`
}

type Token struct {
	Token string `json:"token"`
}

type chirpyRedWebhookData struct {
	UserID int `json:"user_id"`
}

type chipryRedWebhookBody struct {
	Event string               `json:"event"`
	Data  chirpyRedWebhookData `json:"data"`
}
