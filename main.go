package main // Declares that this file belongs to the 'main' package, which is required for executable programs

import ( // Begins an import block to include external packages
	"encoding/json"
	"fmt" // Imports the fmt package, which provides formatted I/O functions
	"log"
	"net/http" // Imports the standard library's HTTP package, which provides HTTP client and server implementations
	"os"
	"strings"
	"sync/atomic" // Imports the atomic package, which provides low-level atomic memory primitives for synchronization
	"time"

	"chirpy/internal/database"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//Create a struct in main.go that will hold any stateful, in-memory data we'll need to keep track of. In our case, we just need to keep track of the number of requests we've received.
type apiConfig struct {
fileserverHits atomic.Int32
dbQueries      *database.Queries
platform	   string			
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
	html := fmt.Sprintf(`<html>
               <body>
               <h1>Welcome, Chirpy Admin</h1>
               <p>Chirpy has been visited %d times!</p>
               </body>
       </html>`, hits)
	
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("This is not a dev environment, you cannot reset metrics or wipe the users table"))
		return
	}
    cfg.fileserverHits.Store(0)
	
	err := cfg.dbQueries.DeleteUsers(req.Context())
	if err != nil {
		log.Printf("Error deleting users: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not delete users")
		return
	}
	w.WriteHeader(http.StatusOK)
    w.Write([]byte("Metrics hits reset to 0, and all users have been deleted from the db"))

}

// ---- Request / Response Types ----

type chirpRequest struct {
    Body string `json:"body"`
	UserID string `json:"user_id"`
}

type errorResponse struct {
    Error string `json:"error"`
}

type chirpResponse struct {
    ID        string   `json:"id"`
    CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body	  string	`json:"body"`
	UserID	  string    `json:"user_id"`
}

type userResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

// ---- Response Helpers ----

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)

    if err := json.NewEncoder(w).Encode(payload); err != nil {
        // Fallback if encoding fails
        http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
    }
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
    respondWithJSON(w, code, errorResponse{Error: msg})
}

func cleanProfanity(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if lowerWord == "kerfuffle" || lowerWord == "sharbert" || lowerWord == "fornax" {
			words[i] = "****" // We are acting on the actual index of a slice from words here, word is just a copy of the slice as per range values
		}
	}
	return strings.Join(words, " ")
}


// ---- Chirp Handler ----

func (cfg *apiConfig) ChirpHandler(w http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()

    var chirp chirpRequest
	
    if err := json.NewDecoder(req.Body).Decode(&chirp); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
        return
    }

    if len(chirp.Body) == 0 {
        respondWithError(w, http.StatusBadRequest, "Body cannot be empty")
        return
    }

    if len(chirp.Body) > 140 {
        respondWithError(w, http.StatusBadRequest, "Chirp is too long (max 140 characters)")
        return
    }

	cleaned := cleanProfanity(chirp.Body)
	
	userID, err := uuid.Parse(chirp.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	chirpDB, err := cfg.dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{Body: cleaned, UserID: userID})
	if err != nil {
		log.Printf("Error publishing Chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not publish the Chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, chirpResponse{
    	ID: 		chirpDB.ID.String(),
    	CreatedAt:	chirpDB.CreatedAt,
		UpdatedAt:  chirpDB.UpdatedAt,
		Body: 		chirpDB.Body,
		UserID: 	chirpDB.UserID.String(),
	})
}

// ---- GetChirps Handler ----

func (cfg *apiConfig) GetChirpsHandler(w http.ResponseWriter, req *http.Request) {
    defer req.Body.Close()

	chirps, err := cfg.dbQueries.GetChirps(req.Context())
	if err != nil {
		log.Printf("Error fetching chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not fetch chirps")
		return
	}

	response := make([]chirpResponse, 0, len(chirps))
	for _, c := range chirps {
    response = append(response, chirpResponse{
        ID:        c.ID.String(),
        CreatedAt: c.CreatedAt,
        UpdatedAt: c.UpdatedAt,
        Body:      c.Body,
        UserID:    c.UserID.String(),
    })
	}
	respondWithJSON(w, http.StatusOK, response)
}


// ---------- Create Db User -----------

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "JSON Body is empty or malformed")
		return
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, userResponse{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func main() { // Defines the main function, which is the entry point of the Go program
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := &apiConfig{
		dbQueries: database.New(db),
		platform: platform,
	}
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

	mux.HandleFunc("POST /api/chirps", apiCfg.ChirpHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirpsHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUser)

	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	server := &http.Server{ // Creates a new HTTP server struct and configures its fields
		Addr:    ":8080", // Sets the network address to listen on (port 8080 on all interfaces, e.g., localhost:8080)
		Handler: mux,     // Assigns the ServeMux as the handler that will process all incoming requests
	}
	server.ListenAndServe() // Starts the server and blocks forever, listening for and handling incoming HTTP connections
}
