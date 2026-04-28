package viewer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	allowedOrigins := []string{"https://app.studyclaw.com", "https://another.com"}
	srv := NewServer(8080, allowedOrigins)

	// Pre-populate store
	srv.StoreContent("test-id", "Test Title", "mermaid", "graph TD; A-->B")

	mux := srv.setupMux()

	tests := []struct {
		name           string
		origin         string
		expectedHeader string
	}{
		{
			name:           "Allowed origin",
			origin:         "https://app.studyclaw.com",
			expectedHeader: "https://app.studyclaw.com",
		},
		{
			name:           "Another allowed origin",
			origin:         "https://another.com",
			expectedHeader: "https://another.com",
		},
		{
			name:           "Disallowed origin",
			origin:         "https://malicious.com",
			expectedHeader: "",
		},
		{
			name:           "No origin",
			origin:         "",
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/content/test-id", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status OK, got %v", rr.Code)
			}

			got := rr.Header().Get("Access-Control-Allow-Origin")
			if got != tt.expectedHeader {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedHeader, got)
			}

			if tt.expectedHeader != "" {
				vary := rr.Header().Get("Vary")
				if vary != "Origin" {
					t.Errorf("expected Vary: Origin, got %q", vary)
				}
			}
		})
	}
}

func TestCORSOnErrorResponses(t *testing.T) {
	allowedOrigins := []string{"https://app.studyclaw.com"}
	srv := NewServer(8080, allowedOrigins)
	mux := srv.setupMux()

	tests := []struct {
		name           string
		path           string
		origin         string
		expectedStatus int
		expectCORS     bool
	}{
		{
			name:           "Missing id with allowed origin gets CORS header",
			path:           "/api/content/",
			origin:         "https://app.studyclaw.com",
			expectedStatus: http.StatusBadRequest,
			expectCORS:     true,
		},
		{
			name:           "Not found with allowed origin gets CORS header",
			path:           "/api/content/nonexistent",
			origin:         "https://app.studyclaw.com",
			expectedStatus: http.StatusNotFound,
			expectCORS:     true,
		},
		{
			name:           "Missing id with disallowed origin gets no CORS header",
			path:           "/api/content/",
			origin:         "https://malicious.com",
			expectedStatus: http.StatusBadRequest,
			expectCORS:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			got := rr.Header().Get("Access-Control-Allow-Origin")
			if tt.expectCORS {
				if got != tt.origin {
					t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.origin, got)
				}
				vary := rr.Header().Get("Vary")
				if vary != "Origin" {
					t.Errorf("expected Vary: Origin, got %q", vary)
				}
			} else {
				if got != "" {
					t.Errorf("expected no Access-Control-Allow-Origin header, got %q", got)
				}
			}
		})
	}
}
