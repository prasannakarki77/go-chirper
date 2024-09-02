package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits uint64
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&cfg.fileserverHits, 1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))

	apiCfg := &apiConfig{}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fs)))

	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", atomic.LoadUint64(&apiCfg.fileserverHits))
	})

	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		atomic.StoreUint64(&apiCfg.fileserverHits, 0)
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
			return
		}

		resp := struct {
			Valid bool   `json:"valid,omitempty"`
			Error string `json:"error,omitempty"`
		}{Valid: len(params.Body) <= 10}

		if !resp.Valid {
			resp.Error = "Chirp is too long"
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(resp)
	})

	http.ListenAndServe(":8080", mux)
}
