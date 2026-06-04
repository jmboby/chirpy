package main // Declares that this file belongs to the 'main' package, which is required for executable programs

import ( // Begins an import block to include external packages
	"net/http" // Imports the standard library's HTTP package, which provides HTTP client and server implementations
	"sync/atomic" // Imports the atomic package, which provides low-level atomic memory primitives for synchronization
	"fmt" // Imports the fmt package, which provides formatted I/O functions
)

//Create a struct in main.go that will hold any stateful, in-memory data we'll need to keep track of. In our case, we just need to keep track of the number of requests we've received.
type apiConfig struct {
fileserverHits atomic.Int32
}

// write a new middleware method on a *apiConfig that increments the fileserverHits counter every time it's called.
// The atomic.Int32 type has an .Add() method, use it to safely increment the number of fileserverHits.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	hits := cfg.fileserverHits.Load()
	html := fmt.Sprintf(
		"<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>",
		hits,
	)
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    
    cfg.fileserverHits.Store(0)
    w.Write([]byte("Hits reset to 0\n"))
}

func main() { // Defines the main function, which is the entry point of the Go program
	apiCfg := &apiConfig{}
	mux := http.NewServeMux() // Creates a new HTTP request multiplexer (router) that matches incoming requests against registered handlers
	// Handler (noun) = an object that implements the http.Handler interface (has a ServeHTTP method)
	// Handle (verb) = the method used to register a handler for a specific URL pattern
	 // Wrap the handler - middleware counts the hit, then calls the file server
    mux.Handle("/app/", apiCfg.middlewareMetricsInc(
        http.StripPrefix("/app/", http.FileServer(http.Dir("."))),
    ))
	//mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(".")))) // Registers a file server handler on the root path "/" that serves files from the current directory
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	server := &http.Server{ // Creates a new HTTP server struct and configures its fields
		Addr:    ":8080", // Sets the network address to listen on (port 8080 on all interfaces, e.g., localhost:8080)
		Handler: mux,     // Assigns the ServeMux as the handler that will process all incoming requests
	}
	server.ListenAndServe() // Starts the server and blocks forever, listening for and handling incoming HTTP connections
}
