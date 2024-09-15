package database

import (
	"encoding/json"
	"io"
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

	file, err := os.Open(db.path)

	if err != nil {
		return Chirp{}, err
	}
	defer file.Close()

	var dbVal DBStructure

	fileVal, err := io.ReadAll(file)

	if err != nil {
		return Chirp{}, err
	}
	err = json.Unmarshal(fileVal, &dbVal)

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

	updatedFileVal, err := json.Marshal(dbVal)

	if err != nil {
		return Chirp{}, err
	}

	err = os.WriteFile(db.path, updatedFileVal, 0644)
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
	err = os.WriteFile(db.path, updatedFileVal, 0644)
	if err != nil {
		return err
	}

	return nil
}
