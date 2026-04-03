package visual

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/roshan30-git/picoclaw-scholar/pkg/viewer"
)

type Manager struct {
	server *viewer.Server
}

func NewManager(server *viewer.Server) *Manager {
	return &Manager{server: server}
}

// RegisterVisual takes raw visual code, stores it in the Viewer server, and returns the ID.
func (m *Manager) RegisterVisual(title string, vType string, source string) string {
	b := make([]byte, 8)
	rand.Read(b)
	id := hex.EncodeToString(b)
	
	m.server.StoreContent(id, title, vType, source)
	return id
}

// GenerateCircuit creates an SVG and registers it immediately.
func (m *Manager) GenerateCircuit(title string, component string) string {
	svg := GenerateCircuitSVG(component)
	return m.RegisterVisual(title, "circuit", svg)
}
