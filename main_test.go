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

func TestValidateChirpHandler(t *testing.T) {
	apiCfg := &apiConfig{}

	body := `{"body":"This is a kerfuffle opinion"}`

	req := httptest.NewRequest("POST", "/api/validate_chirp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	//Call the handler directly with the write arg being rec and req our request above
	apiCfg.validateChirpHandler(rec, req)

	// Check status code
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp chirpResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.CleanedBody != "This is a **** opinion" {
		t.Errorf(("expected %q, got %q"), "This is a **** opinion", resp.CleanedBody)
	}
}