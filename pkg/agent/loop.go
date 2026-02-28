package agent

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	"github.com/roshan30-git/picoclaw-scholar/pkg/channels"
	"github.com/roshan30-git/picoclaw-scholar/pkg/config"
	"github.com/roshan30-git/picoclaw-scholar/pkg/memory"
	"github.com/roshan30-git/picoclaw-scholar/pkg/study"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
	"github.com/roshan30-git/picoclaw-scholar/pkg/visual"
)

type AgentLoop struct {
	cfg     *config.Config
	bus     *bus.MessageBus
	provider tools.LLMProvider
	tools      map[string]tools.Tool
	mgr        *channels.Manager
	visManager  *visual.Manager
	calendar    *study.CalendarEngine
	reflections *memory.ReflectionManager
	inbox       chan bus.InboundMessage
	quit        chan struct{}
	sessions    map[string][]tools.Message
}

func NewAgentLoop(cfg *config.Config, b *bus.MessageBus, provider tools.LLMProvider, vm *visual.Manager, cal *study.CalendarEngine, mem *memory.ReflectionManager) *AgentLoop {
	return &AgentLoop{
		cfg:         cfg,
		bus:         b,
		provider:    provider,
		tools:       make(map[string]tools.Tool),
		visManager:  vm,
		calendar:    cal,
		reflections: mem,
		inbox:      b.Subscribe(),
		quit:       make(chan struct{}),
		sessions:   make(map[string][]tools.Message),
	}
}

func (l *AgentLoop) SetChannelManager(mgr *channels.Manager) {
	l.mgr = mgr
}

func (l *AgentLoop) RegisterTool(t tools.Tool) {
	l.tools[t.Name()] = t
	log.Printf("Registered tool: %s", t.Name())
}

func (l *AgentLoop) Run(ctx context.Context) error {
	log.Println("Agent loop running...")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-l.quit:
			return nil
		case msg := <-l.inbox:
			l.handleMessage(ctx, msg)
		}
	}
}

func (l *AgentLoop) Stop() {
	close(l.quit)
}

func (l *AgentLoop) handleMessage(ctx context.Context, msg bus.InboundMessage) {
	log.Printf("[%s] Message from %s", msg.Channel, msg.From)

	sessionID := msg.Channel + ":" + msg.ChatID
	history := l.sessions[sessionID]

	// Naive correction detection MVP
	lowerMsg := strings.ToLower(msg.Content)
	if (strings.Contains(lowerMsg, "wrong") || strings.Contains(lowerMsg, "incorrect") || strings.Contains(lowerMsg, "actually")) && len(history) > 0 {
		lastMsg := history[len(history)-1]
		if lastMsg.Role == "model" && l.reflections != nil {
			log.Println("[AgentLoop] Correction detected! Logging reflection.")
			l.reflections.LogMistake(msg.Content, lastMsg.Content)
		}
	}

	enrichedContent := msg.Content
	
	var contextPreamble string
	if l.calendar != nil {
		contextPreamble += l.calendar.GetContext() + "\n\n"
	}
	if l.reflections != nil {
		lessons := l.reflections.GetRecentReflections()
		if lessons != "" {
			contextPreamble += lessons + "\n\n"
		}
	}
	
	if contextPreamble != "" {
		enrichedContent = contextPreamble + "User Message:\n" + msg.Content
	}

	history = append(history, tools.Message{Role: "user", Content: enrichedContent})

	var toolDefs []tools.ToolDefinition
	for _, t := range l.tools {
		toolDefs = append(toolDefs, tools.ToolDefinition{
			Type: "function",
			Function: tools.ToolFunctionDefinition{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.Parameters(),
			},
		})
	}

	resp, err := l.provider.Chat(ctx, history, toolDefs, l.cfg.ModelName, nil)
	if err != nil {
		log.Printf("LLM error: %v", err)
		return
	}

	if len(resp.ToolCalls) > 0 {
		history = append(history, tools.Message{Role: "model", Content: resp.Content})
		for _, tc := range resp.ToolCalls {
			log.Printf("Executing tool: %s", tc.Name)
			if t, ok := l.tools[tc.Name]; ok {
				res := t.Execute(ctx, tc.Args)
				resultMsg := fmt.Sprintf("Tool %s completed. Result: %s", tc.Name, res.ForLLM)
				history = append(history, tools.Message{Role: "user", Content: resultMsg})
			} else {
				history = append(history, tools.Message{Role: "user", Content: fmt.Sprintf("Tool %s not found", tc.Name)})
			}
		}

		resp, err = l.provider.Chat(ctx, history, toolDefs, l.cfg.ModelName, nil)
		if err != nil {
			log.Printf("LLM error after tool call: %v", err)
			return
		}
	}

	history = append(history, tools.Message{Role: "model", Content: resp.Content})
	if len(history) > 20 {
		history = history[len(history)-20:]
	}
	l.sessions[sessionID] = history

	// Parse visual tags securely before sending
	out := bus.OutboundMessage{
		ChatID:  msg.ChatID,
		Content: resp.Content,
		Channel: msg.Channel,
	}

	if l.visManager != nil {
		diagramRe := regexp.MustCompile(`(?s)<diagram>(.*?)</diagram>`)
		formulaRe := regexp.MustCompile(`(?s)<formula>(.*?)</formula>`)
		circuitRe := regexp.MustCompile(`(?s)<circuit>(.*?)</circuit>`)

		if matches := diagramRe.FindStringSubmatch(out.Content); len(matches) > 1 {
			out.Content = diagramRe.ReplaceAllString(out.Content, "*(Diagram generated ✨)*")
			out.VisualID = l.visManager.RegisterVisual("AI Diagram", "mermaid", matches[1])
			out.VisualType = "mermaid"
		} else if matches := formulaRe.FindStringSubmatch(out.Content); len(matches) > 1 {
			out.Content = formulaRe.ReplaceAllString(out.Content, "*(Formula generated ✨)*")
			out.VisualID = l.visManager.RegisterVisual("AI Formula", "formula", matches[1])
			out.VisualType = "formula"
		} else if matches := circuitRe.FindStringSubmatch(out.Content); len(matches) > 1 {
			out.Content = circuitRe.ReplaceAllString(out.Content, "*(Circuit generated ✨)*")
			// The circuit tag content is just the component name (e.g., "resistor")
			out.VisualID = l.visManager.GenerateCircuit("AI Circuit", matches[1])
			out.VisualType = "circuit"
		}
	}

	if l.mgr != nil && resp.Content != "" {
		if err := l.mgr.Send(ctx, out); err != nil {
			log.Printf("Failed to route response: %v", err)
		}
	}

	log.Printf("[%s] Response to %s sent", msg.Channel, msg.ChatID)
}
