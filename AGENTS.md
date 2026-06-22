# Coaching Mode

You are a Go coach for this project. Never write code for me. Instead, give hints, explanations, and pointers to docs or patterns in the existing codebase.

# Chirpy Project Wiring

## Request → Response flow

```
HTTP request → ServeMux (main.go) → handler method on *apiConfig → database.Queries → JSON response
```

- `main.go` registers routes via `mux.HandleFunc("METHOD /path", handler)` at the bottom of `main()`.
- Handlers are methods on `*apiConfig` (e.g., `func (cfg *apiConfig) createUser(...)`).
- `apiConfig` holds shared state: `dbQueries *database.Queries` and `fileserverHits atomic.Int32`.

## JSON handling

- **Request bodies**: decoded with `json.NewDecoder(r.Body).Decode(&struct)`. Go's decoder silently ignores unknown JSON keys — validate required fields after decoding.
- **Response bodies**: encoded with `json.NewEncoder(w).Encode(payload)`, wrapped in `respondWithJSON(w, code, payload)`.
- Structs with `json:"field_name"` tags control JSON key names (see `userResponse`, `errorResponse`, etc.).

## Middleware

- `middlewareMetricsInc` wraps a handler, increments the hit counter, then calls the next handler.
- Pattern: `mux.Handle("/path", cfg.middlewareMetricsInc(nextHandler))`.

## Database

- Uses `database.Queries` from `internal/database` package, initialized in `main()` from a PostgreSQL connection string (`DB_URL` env var).
- Queries are methods on `*database.Queries`, take `context.Context` and parameters.

# Key Go patterns used in this repo

- Handler methods on a config struct (`*apiConfig`) for dependency injection (no globals).
- `atomic.Int32` for safe concurrent counter.
- `json.Decoder` / `json.NewEncoder` for streaming JSON I/O.
- `http.ServeMux` with method-based routing (`"GET /path"`, `"POST /path"`).
