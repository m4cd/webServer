package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	database := DB{path: path, mux: &sync.RWMutex{}}
	err := database.ensureDB()
	return &database, err
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbContents, _ := db.loadDB()
	dbChirps, err := db.GetChirps()

	if err != nil {
		fmt.Println("Error loading DB in CreateChirp: ", err)
	}

	// find max id
	max := 0
	for _, c := range dbChirps {
		if c.ID > max {
			max = c.ID
		}
	}

	newID := max + 1
	chirp := Chirp{
		ID:   newID,
		Body: body,
	}

	dbChirps = append(dbChirps, chirp)

	dbChirpsMap := make(map[int]Chirp)

	for _, c := range dbChirps {
		ch := Chirp{
			ID:   c.ID,
			Body: c.Body,
		}
		dbChirpsMap[c.ID] = ch
	}

	dbChirpsStruct := dbContents
	dbChirpsStruct.Chirps = dbChirpsMap

	db.writeDB(dbChirpsStruct)

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	dbChirps, err := db.loadDB()
	if err == nil {
		mapChirps := []Chirp{}

		for _, c := range dbChirps.Chirps {
			mapChirps = append(mapChirps, c)
		}

		return mapChirps, nil
	}
	return []Chirp{}, err
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if err != nil {
		//e := os.WriteFile(db.path, []byte{}, 0664)
		e := db.writeDB(DBStructure{})
		return e
	}
	return err
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {

	fileExists := db.ensureDB()
	if fileExists == nil {
		db.mux.Lock()
		dbContents, _ := ioutil.ReadFile(db.path)
		db.mux.Unlock()

		var dbChirps DBStructure

		err := json.Unmarshal(dbContents, &dbChirps)
		if err != nil {
			fmt.Println("Error while decoding")
			return DBStructure{}, err
		}

		return dbChirps, nil
	}
	return DBStructure{}, fileExists

}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	data, err := json.Marshal(dbStructure)
	if err == nil {
		db.mux.Lock()
		e := os.WriteFile(db.path, data, 0664)
		db.mux.Unlock()
		return e
	}
	return err
}

// getUsers returns all userss in the database
func (db *DB) getUsers() ([]User, error) {
	dbUsers, err := db.loadDB()
	if err == nil {
		mapUsers := []User{}

		for _, u := range dbUsers.Users {
			mapUsers = append(mapUsers, u)
		}

		return mapUsers, nil
	}
	return []User{}, err
}

func (db *DB) CreateUser(email string) (User, error) {
	dbContents, _ := db.loadDB()
	dbUsers, err := db.getUsers()

	if err != nil {
		fmt.Println("Error while loading db")
	}

	// find max id
	max := 0
	for _, u := range dbContents.Users {
		if u.ID > max {
			max = u.ID
		}
	}

	newID := max + 1
	user := User{
		ID:    newID,
		Email: email,
	}

	dbUsers = append(dbUsers, user)
	dbUsersMap := make(map[int]User)

	for _, u := range dbUsers {
		us := User{
			ID:    u.ID,
			Email: u.Email,
		}
		dbUsersMap[u.ID] = us
	}
	dbUsersStruct := dbContents
	dbUsersStruct.Users = dbUsersMap

	db.writeDB(dbUsersStruct)

	return user, nil
}
