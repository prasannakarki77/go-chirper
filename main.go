package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/", fs)

	mux.HandleFunc("/healthz", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; chartset=utf-8")
		res.Header().Set("X-Hello-Bootdev", "Showing these being set")
		res.WriteHeader(200)
		res.Write([]byte("OK"))
	})

	serv := &http.Server{Handler: mux, Addr: ":8080"}
	serv.ListenAndServe()

}
