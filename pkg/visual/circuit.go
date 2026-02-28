package visual

import (
	"fmt"
	"strings"
)

// GenerateCircuitSVG returns a rudimentary SVG string representing an electrical circuit component
// based on a simple text description.
func GenerateCircuitSVG(componentType string) string {
	componentType = strings.ToLower(componentType)
	
	switch {
	case strings.Contains(componentType, "resistor"):
		return `<svg width="200" height="100" viewBox="0 0 200 100" xmlns="http://www.w3.org/2000/svg">
			<line x1="20" y1="50" x2="60" y2="50" stroke="black" stroke-width="4"/>
			<polyline points="60,50 65,35 75,65 85,35 95,65 105,35 115,65 125,50 140,50" fill="none" stroke="black" stroke-width="4"/>
			<line x1="140" y1="50" x2="180" y2="50" stroke="black" stroke-width="4"/>
			<text x="100" y="80" text-anchor="middle" font-family="monospace" font-size="16">Resistor</text>
		</svg>`
	case strings.Contains(componentType, "capacitor"):
		return `<svg width="200" height="100" viewBox="0 0 200 100" xmlns="http://www.w3.org/2000/svg">
			<line x1="20" y1="50" x2="90" y2="50" stroke="black" stroke-width="4"/>
			<line x1="90" y1="20" x2="90" y2="80" stroke="black" stroke-width="4"/>
			<line x1="110" y1="20" x2="110" y2="80" stroke="black" stroke-width="4"/>
			<line x1="110" y1="50" x2="180" y2="50" stroke="black" stroke-width="4"/>
			<text x="100" y="95" text-anchor="middle" font-family="monospace" font-size="16">Capacitor</text>
		</svg>`
	case strings.Contains(componentType, "inductor"):
		return `<svg width="200" height="100" viewBox="0 0 200 100" xmlns="http://www.w3.org/2000/svg">
			<line x1="20" y1="50" x2="60" y2="50" stroke="black" stroke-width="4"/>
			<path d="M 60 50 Q 75 20 90 50 Q 105 20 120 50 Q 135 20 150 50" fill="none" stroke="black" stroke-width="4"/>
			<line x1="150" y1="50" x2="180" y2="50" stroke="black" stroke-width="4"/>
			<text x="100" y="80" text-anchor="middle" font-family="monospace" font-size="16">Inductor</text>
		</svg>`
	default:
		return fmt.Sprintf(`<svg width="200" height="100" viewBox="0 0 200 100" xmlns="http://www.w3.org/2000/svg">
			<rect x="50" y="30" width="100" height="40" fill="none" stroke="black" stroke-width="4"/>
			<text x="100" y="55" text-anchor="middle" font-family="monospace" font-size="16">%s</text>
		</svg>`, componentType)
	}
}
