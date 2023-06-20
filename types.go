package main

type apiConfig struct {
	fileserverHits int
}

type errorParameters struct {
	Error string `json:"error"`
}

type bodyParameters struct {
	BodyJSON string `json:"body"`
}

type cleanedBody struct {
	CleanedBody string `json:"cleaned_body"`
}

type bodyUser struct {
	Email string `json:"email"`
}
