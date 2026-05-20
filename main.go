package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	
	// Create a new Server and configure it
	server := &http.Server{
    		Addr:    ":8080",
    		Handler: mux,
	}
	server.ListenAndServe()	
}