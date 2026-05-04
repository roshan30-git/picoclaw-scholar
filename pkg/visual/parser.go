package visual

import (
	"regexp"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
)

// Parser handles the detection and conversion of visual/interactive tags in LLM responses.
type Parser struct {
	manager *Manager
}

func NewParser(m *Manager) *Parser {
	return &Parser{manager: m}
}

// ParseResult represents the result of parsing visual tags.
type ParseResult struct {
	CleanContent string
	VisualID     string
	VisualType   string
}

var (
	diagramRe = regexp.MustCompile(`(?s)<diagram>(.*?)</diagram>`)
	formulaRe = regexp.MustCompile(`(?s)<formula>(.*?)</formula>`)
	circuitRe = regexp.MustCompile(`(?s)<circuit>(.*?)</circuit>`)
)

// ParseContent extracts <diagram>, <formula>, and <circuit> tags from a message and interacts with the visual manager.
func (p *Parser) ParseContent(content string) ParseResult {
	if p.manager == nil {
		return ParseResult{CleanContent: content}
	}

	res := ParseResult{CleanContent: content}

	if matches := diagramRe.FindStringSubmatch(content); len(matches) > 1 {
		res.CleanContent = diagramRe.ReplaceAllString(content, "*(Diagram generated ✨)*")
		res.VisualID = p.manager.RegisterVisual("AI Diagram", "mermaid", matches[1])
		res.VisualType = "mermaid"
	} else if matches := formulaRe.FindStringSubmatch(content); len(matches) > 1 {
		res.CleanContent = formulaRe.ReplaceAllString(content, "*(Formula generated ✨)*")
		res.VisualID = p.manager.RegisterVisual("AI Formula", "formula", matches[1])
		res.VisualType = "formula"
	} else if matches := circuitRe.FindStringSubmatch(content); len(matches) > 1 {
		res.CleanContent = circuitRe.ReplaceAllString(content, "*(Circuit generated ✨)*")
		res.VisualID = p.manager.GenerateCircuit("AI Circuit", matches[1])
		res.VisualType = "circuit"
	}

	return res
}

// ApplyVisuals updates an OutboundMessage based on detected visual tags in its content.
func (p *Parser) ApplyVisuals(out *bus.OutboundMessage) {
	result := p.ParseContent(out.Content)
	out.Content = result.CleanContent
	out.VisualID = result.VisualID
	out.VisualType = result.VisualType
}
