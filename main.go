package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	serv := http.Server{Handler: mux, Addr: ":8080"}
	serv.ListenAndServe()

}
