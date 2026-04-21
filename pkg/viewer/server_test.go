package viewer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	allowedOrigins := []string{"https://trusted.com", "https://app.studyclaw.com"}
	srv := NewServer(8080, allowedOrigins)
	srv.StoreContent("test-id", "Test Title", "mermaid", "graph TD; A-->B")

	mux := srv.setupMux()

	tests := []struct {
		name           string
		origin         string
		expectedHeader string
	}{
		{
			name:           "Allowed Origin 1",
			origin:         "https://trusted.com",
			expectedHeader: "https://trusted.com",
		},
		{
			name:           "Allowed Origin 2",
			origin:         "https://app.studyclaw.com",
			expectedHeader: "https://app.studyclaw.com",
		},
		{
			name:           "Unauthorized Origin",
			origin:         "https://malicious.com",
			expectedHeader: "",
		},
		{
			name:           "No Origin Header",
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

			gotHeader := rr.Header().Get("Access-Control-Allow-Origin")
			if gotHeader != tt.expectedHeader {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedHeader, gotHeader)
			}
		})
	}
}
