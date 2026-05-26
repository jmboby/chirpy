package main // Declares that this file belongs to the 'main' package, which is required for executable programs

import ( // Begins an import block to include external packages
	"net/http" // Imports the standard library's HTTP package, which provides HTTP client and server implementations
)

func main() { // Defines the main function, which is the entry point of the Go program
	mux := http.NewServeMux() // Creates a new HTTP request multiplexer (router) that matches incoming requests against registered handlers
	// Handler (noun) = an object that implements the http.Handler interface (has a ServeHTTP method)
	// Handle (verb) = the method used to register a handler for a specific URL pattern
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(".")))) // Registers a file server handler on the root path "/" that serves files from the current directory
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{ // Creates a new HTTP server struct and configures its fields
		Addr:    ":8080", // Sets the network address to listen on (port 8080 on all interfaces, e.g., localhost:8080)
		Handler: mux,     // Assigns the ServeMux as the handler that will process all incoming requests
	}
	server.ListenAndServe() // Starts the server and blocks forever, listening for and handling incoming HTTP connections
}
