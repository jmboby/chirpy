package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)	

func TestCleanProfanity(t *testing.T) {
	tests := []struct {
		input	string
		expected 	string
	}{
		{"kerfuffle", "****"},
		{"sharbert", "****"},
		{"fornax", "****"},
	}
	for _, tt := range tests {
        result := cleanProfanity(tt.input)
        if result != tt.expected {
            t.Errorf("cleanProfanity(%q) = %q, want %q", tt.input, result, tt.expected)
        }
    }
}

func TestChirpHandler(t *testing.T) {
	apiCfg := &apiConfig{}

	body := `{"body":"This is a kerfuffle opinion", "user_id":"123e4567-e89b-12d3-a456-426614174000"}`

	req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	//Call the handler directly with the write arg being rec and req our request above
	apiCfg.ChirpHandler(rec, req)

	// Check status code
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp chirpResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Body != "This is a **** opinion" {
		t.Errorf(("expected %q, got %q"), "This is a **** opinion", resp.Body)
	}
}