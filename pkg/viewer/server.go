package viewer

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	port int
}

func NewServer(port int) *Server {
	return &Server{port: port}
}

func (s *Server) Start() error {
	fs := http.FileServer(http.Dir("./pkg/viewer/static"))
	http.Handle("/", fs)
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Viewer server listening on http://127.0.0.1%s", addr)
	return http.ListenAndServe(addr, nil)
}
