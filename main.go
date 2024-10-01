package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prasannakarki77/go-chirper/internal/database"
)

type apiConfig struct {
	fileserverHits int
	token          string
}

type ErrorRes struct {
	Error string `json:"error"`
}

var profaneWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits = cfg.fileserverHits + 1
		next.ServeHTTP(w, r)
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorRes{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func replaceProfaneWords(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		cleanWord := strings.ToLower(strings.Trim(word, ".,!?"))
		for _, badWord := range profaneWords {
			if cleanWord == badWord {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
		log.Println("Falling back to OS environment variables")
	}

	// Get token from environment
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable not set")
	}
	fmt.Println("token:", token)
	apiCfg := &apiConfig{
		fileserverHits: 0,
		token:          token,
	}
	db, err := database.NewDB("database.json")

	if err != nil {
		fmt.Println(err)
		return
	}

	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fs)))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))

	})

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1>   <p>Chirpy has been visited %d times!</p></body></html>", apiCfg.fileserverHits)))
	})

	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileserverHits = 0
		w.WriteHeader(200)
	})

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}

		var params parameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		if len(params.Body) > 140 {
			w.WriteHeader(http.StatusBadRequest)
			respondWithError(w, 400, "Chirp is too long")
			return
		}

		cleanedBody := replaceProfaneWords(params.Body)

		chirp, err := db.CreateChirp(cleanedBody)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithJSON(w, http.StatusCreated, chirp)
	})

	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {

		chirps, err := db.GetChirps()

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			respondWithError(w, 400, "failed")
		}
		respondWithJSON(w, http.StatusOK, chirps)

	})
	mux.HandleFunc("/api/chirps/{chirpId}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("chirpId")

		chirpId, err := strconv.Atoi(id)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
			return
		}

		// Fetch the chirp from the database using the chirpId
		chirp, err := db.GetChirpById(chirpId)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}

		respondWithJSON(w, http.StatusOK, chirp)
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {

		type Params struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var params Params
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		user, err := db.CreateUser(params.Email, params.Password)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, user)
	})

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {

		type Params struct {
			Email            string `json:"email"`
			Password         string `json:"password"`
			ExpiresInSeconds string `json:"expires_in_seconds"`
		}

		var params Params
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		user, err := db.LoginUser(params.Email, params.Password)

		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"user":  user,
			"token": token,
		})
	})

	serv := &http.Server{Handler: mux, Addr: ":8080"}
	serv.ListenAndServe()

}
