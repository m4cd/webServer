package webapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	api := DB{path: path, mux: &sync.RWMutex{}}
	err := api.ensureDB()
	return &api, err
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
