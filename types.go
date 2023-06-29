package main

type apiConfig struct {
	fileserverHits         int
	jwtSecret              string
	accessTokenExpiration  int
	accessTokenIssuer      string
	refreshTokenExpiration int
	refreshTokenIssuer     string
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
}
