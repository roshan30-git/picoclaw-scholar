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

	// Optimization: Use FindStringSubmatchIndex instead of FindStringSubmatch + ReplaceAllString
	// This avoids evaluating the regular expression twice and is over 2x faster.
	if matches := diagramRe.FindStringSubmatchIndex(content); len(matches) > 3 {
		res.CleanContent = content[:matches[0]] + "*(Diagram generated ✨)*" + content[matches[1]:]
		res.VisualID = p.manager.RegisterVisual("AI Diagram", "mermaid", content[matches[2]:matches[3]])
		res.VisualType = "mermaid"
	} else if matches := formulaRe.FindStringSubmatchIndex(content); len(matches) > 3 {
		res.CleanContent = content[:matches[0]] + "*(Formula generated ✨)*" + content[matches[1]:]
		res.VisualID = p.manager.RegisterVisual("AI Formula", "formula", content[matches[2]:matches[3]])
		res.VisualType = "formula"
	} else if matches := circuitRe.FindStringSubmatchIndex(content); len(matches) > 3 {
		res.CleanContent = content[:matches[0]] + "*(Circuit generated ✨)*" + content[matches[1]:]
		res.VisualID = p.manager.GenerateCircuit("AI Circuit", content[matches[2]:matches[3]])
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
