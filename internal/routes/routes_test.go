package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kosis/internal/config"
	"gorm.io/gorm"
)

func TestCORS(t *testing.T) {
	// Create a mock DB and config
	db := &gorm.DB{}
	cfg := &config.Config{
		AllowedOrigins: "http://localhost:3000, https://example.com ",
	}

	router := SetupRouter(db, cfg)

	tests := []struct {
		name           string
		origin         string
		expectedStatus int
		expectedOrigin string
	}{
		{
			name:           "Allowed Origin",
			origin:         "http://localhost:3000",
			expectedStatus: http.StatusOK,
			expectedOrigin: "http://localhost:3000",
		},
		{
			name:           "Another Allowed Origin with whitespace in config",
			origin:         "https://example.com",
			expectedStatus: http.StatusOK,
			expectedOrigin: "https://example.com",
		},
		{
			name:           "Disallowed Origin",
			origin:         "http://malicious.com",
			expectedStatus: http.StatusOK,
			expectedOrigin: "", // Should not be set
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/health", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			originHeader := w.Header().Get("Access-Control-Allow-Origin")
			if originHeader != tc.expectedOrigin {
				t.Errorf("Expected Origin %q, got %q", tc.expectedOrigin, originHeader)
			}
		})
	}
}

func TestCORSWildcard(t *testing.T) {
	// Create a mock DB and config
	db := &gorm.DB{}
	cfg := &config.Config{
		AllowedOrigins: "*",
	}

	router := SetupRouter(db, cfg)

	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://malicious.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	originHeader := w.Header().Get("Access-Control-Allow-Origin")
	if originHeader != "http://malicious.com" { // It should reflect the origin now when wildcard is used
		t.Errorf("Expected Origin 'http://malicious.com', got %q", originHeader)
	}
}
