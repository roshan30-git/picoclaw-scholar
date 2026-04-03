package viewer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

type VisualContent struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`   // 'mermaid', 'formula', 'circuit'
	Source string `json:"source"` // The raw mermaid string, katex string, or SVG code
}

type Server struct {
	port  int
	mu    sync.RWMutex
	store map[string]VisualContent
}

func NewServer(port int) *Server {
	return &Server{
		port:  port,
		store: make(map[string]VisualContent),
	}
}

// StoreContent saves visual data to memory to be fetched by the Mini App.
func (s *Server) StoreContent(id, title, vType, source string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[id] = VisualContent{
		ID:     id,
		Title:  title,
		Type:   vType,
		Source: source,
	}
}

func (s *Server) Start() error {
	fs := http.FileServer(http.Dir("./pkg/viewer/static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/content/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/content/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		s.mu.RLock()
		content, ok := s.store[id]
		s.mu.RUnlock()

		if !ok {
			http.Error(w, "content not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(content)
	})

	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	log.Printf("Viewer server listening on http://%s", addr)
	return http.ListenAndServe(addr, nil)
}
