package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prasannakarki77/go-chirper/internal/database"
)

type apiConfig struct {
	fileserverHits int
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

	apiCfg := &apiConfig{
		fileserverHits: 0,
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

		// cleanedBody := replaceProfaneWords(params.Body)s

		chirp, err := db.CreateChirp(params.Body)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, chirp)
	})

	serv := &http.Server{Handler: mux, Addr: ":8080"}
	serv.ListenAndServe()

}
