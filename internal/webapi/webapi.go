package webapi

import (
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

type Chirps []Chirp

type DBStructure struct {
	Chirps        map[int]Chirp              `json:"chirps"`
	Users         map[int]User               `json:"users"`
	RevokedTokens map[string]jwt.NumericDate `json:"revoked_tokens"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ResponseUser struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type RevokedToken struct {
	TokenString string          `json:"token"`
	RevokeTime  jwt.NumericDate `json:"revoke_time,omitempty"`
}
