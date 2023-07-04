package webapi

import "fmt"

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
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
		ID:       newID,
		Body:     body,
		AuthorID: authorID,
	}

	dbChirps = append(dbChirps, chirp)

	dbChirpsMap := make(map[int]Chirp)

	for _, c := range dbChirps {
		ch := Chirp{
			ID:       c.ID,
			Body:     c.Body,
			AuthorID: c.AuthorID,
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
		mapChirps := make([]Chirp, 0, len(dbChirps.Chirps))

		for _, c := range dbChirps.Chirps {
			mapChirps = append(mapChirps, c)
		}

		return mapChirps, nil
	}
	return []Chirp{}, err
}

func (db *DB) DeleteChirp(id int) error {
	dbChirps, err := db.loadDB()
	if err != nil {
		fmt.Println("Error loading DB:", err)
		return err
	}
	//mapChirps := make([]Chirp, 0, len(dbChirps.Chirps))
	dbChirpsMap := make(map[int]Chirp)

	for _, c := range dbChirps.Chirps {
		if c.ID == id {
			continue
		}
		ch := Chirp{
			ID:       c.ID,
			Body:     c.Body,
			AuthorID: c.AuthorID,
		}
		dbChirpsMap[c.ID] = ch
	}
	dbChirps.Chirps = dbChirpsMap
	db.writeDB(dbChirps)

	return nil
}
