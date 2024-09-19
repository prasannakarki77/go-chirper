package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
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
	fmt.Println(body)
	if err != nil {
		return Chirp{}, err
	}

	dbVal, err := db.loadDB()

	if err != nil {
		return Chirp{}, err
	}

	length := len(dbVal.Chirps)
	fmt.Println(length)

	chirp := Chirp{
		Id:   length + 1,
		Body: body,
	}
	fmt.Println(dbVal)
	if dbVal.Chirps == nil {
		dbVal.Chirps = make(map[int]Chirp)
	}
	dbVal.Chirps[length+1] = chirp
	fmt.Println(dbVal)

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

	if len(fileVal) <= 0 {
		return dbVal, nil
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

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	dbVal, err := db.loadDB()
	if err != nil {
		return []Chirp{}, err
	}

	result := []Chirp{}

	for i := 0; i < len(dbVal.Chirps); i++ {
		result = append(result, dbVal.Chirps[i+1])
	}

	return result, nil

}

func (db *DB) GetChirpById(id int) (Chirp, error) {
	dbVal, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := dbVal.Chirps[id]
	if !ok {
		return Chirp{}, fmt.Errorf("chirp with ID %d not found", id)
	}

	return chirp, nil
}
