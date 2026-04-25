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

			srv.contentHandler(rr, req)

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
