package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/*", http.StripPrefix("/app", fs))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))

	})

	serv := &http.Server{Handler: mux, Addr: ":8080"}
	serv.ListenAndServe()

}
