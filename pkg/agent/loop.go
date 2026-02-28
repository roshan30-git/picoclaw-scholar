package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	"github.com/roshan30-git/picoclaw-scholar/pkg/channels"
	"github.com/roshan30-git/picoclaw-scholar/pkg/config"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
)

type AgentLoop struct {
	cfg     *config.Config
	bus     *bus.MessageBus
	provider tools.LLMProvider
	tools   map[string]tools.Tool
	mgr      *channels.Manager
	inbox    chan bus.InboundMessage
	quit     chan struct{}
	sessions map[string][]tools.Message
}

func NewAgentLoop(cfg *config.Config, b *bus.MessageBus, provider tools.LLMProvider) *AgentLoop {
	return &AgentLoop{
		cfg:      cfg,
		bus:      b,
		provider: provider,
		tools:    make(map[string]tools.Tool),
		inbox:    b.Subscribe(),
		quit:     make(chan struct{}),
		sessions: make(map[string][]tools.Message),
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

	history = append(history, tools.Message{Role: "user", Content: msg.Content})

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

	if l.mgr != nil && resp.Content != "" {
		out := bus.OutboundMessage{
			ChatID:  msg.ChatID,
			Content: resp.Content,
			Channel: msg.Channel,
		}
		if err := l.mgr.Send(ctx, out); err != nil {
			log.Printf("Failed to route response: %v", err)
		}
	}

	log.Printf("[%s] Response to %s sent", msg.Channel, msg.ChatID)
}
