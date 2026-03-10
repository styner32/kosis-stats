package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kosis/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		allowedOrigins string
		reqOrigin      string
		expectedOrigin string
	}{
		{
			name:           "allowed origin matches exactly",
			allowedOrigins: "https://example.com",
			reqOrigin:      "https://example.com",
			expectedOrigin: "https://example.com",
		},
		{
			name:           "allowed origin matches one in list",
			allowedOrigins: "https://foo.com, https://example.com",
			reqOrigin:      "https://example.com",
			expectedOrigin: "https://example.com",
		},
		{
			name:           "origin not in allowed list",
			allowedOrigins: "https://foo.com, https://bar.com",
			reqOrigin:      "https://example.com",
			expectedOrigin: "",
		},
		{
			name:           "no allowed origins specified",
			allowedOrigins: "",
			reqOrigin:      "https://example.com",
			expectedOrigin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				AllowedOrigins: tt.allowedOrigins,
			}

			router := SetupRouter(&gorm.DB{}, cfg)

			req, _ := http.NewRequest(http.MethodOptions, "/health", nil)
			if tt.reqOrigin != "" {
				req.Header.Set("Origin", tt.reqOrigin)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		})
	}
}
