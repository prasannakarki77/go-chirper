package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits = cfg.fileserverHits + 1
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))

	apiCfg := &apiConfig{
		fileserverHits: 0,
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

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}

		type ErrResp struct {
			Error string `json:"error"`
		}
		type ValidResp struct {
			Valid bool `json:"valid"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		fmt.Println(params.Body)
		if err != nil {
			errRes := ErrResp{
				Error: "Something went wrong",
			}
			dat, err := json.Marshal(errRes)
			if err == nil {
				w.Write(dat)
			}
			w.WriteHeader(500)
			return

		}

		if len(params.Body) > 10 {
			errRes := ErrResp{
				Error: "Chirp is too long",
			}
			dat, err := json.Marshal(errRes)
			if err == nil {
				w.Write(dat)
			}
			w.WriteHeader(400)
			return
		}

		validResp := ValidResp{
			Valid: true,
		}
		dat, err := json.Marshal(validResp)
		if err == nil {
			w.Write(dat)
		}
		w.WriteHeader(200)

	})

	serv := &http.Server{Handler: mux, Addr: ":8080"}
	serv.ListenAndServe()

}
