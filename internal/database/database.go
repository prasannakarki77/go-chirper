package database

import (
	"encoding/json"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	id   int
	body string
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// Creates new DB connnection and DB file if it doesn't exists
func NewDB(path string) (*DB, error) {

	mux := &sync.RWMutex{}

	file, err := os.OpenFile(path, os.O_RDWR, 0666)

	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(path)

			if err != nil {
				return nil, os.ErrNotExist
			}
		} else {
			return nil, err
		}
	}
	file.Close()

	db := &DB{
		path: path,
		mux:  mux,
	}
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	err := db.ensureDB()

	if err != nil {
		return Chirp{}, err
	}

	dbVal, err := db.loadDB()

	if err != nil {
		return Chirp{}, err
	}

	length := len(dbVal.Chirps)

	chirp := Chirp{
		id:   length + 1,
		body: body,
	}

	if dbVal.Chirps == nil {
		dbVal.Chirps = make(map[int]Chirp)
	}
	dbVal.Chirps[length] = chirp

	err = db.writeDB(dbVal)

	if err != nil {
		return Chirp{}, nil
	}

	return chirp, nil

}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)

	if err != nil {
		NewDB(db.path)
		return err
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	var dbVal DBStructure

	fileVal, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}
	err = json.Unmarshal(fileVal, &dbVal)
	if err != nil {
		return DBStructure{}, err
	}
	return dbVal, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	updatedFileVal, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, updatedFileVal, 0644)
	if err != nil {
		return err
	}

	return nil
}
