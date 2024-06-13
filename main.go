package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/", fs)

	serv := &http.Server{Handler: mux, Addr: ":8080"}
	serv.ListenAndServe()

}
