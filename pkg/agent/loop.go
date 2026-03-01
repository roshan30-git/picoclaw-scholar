package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
	"github.com/roshan30-git/picoclaw-scholar/pkg/channels"
	"github.com/roshan30-git/picoclaw-scholar/pkg/config"
	pkgdb "github.com/roshan30-git/picoclaw-scholar/pkg/database"
	"github.com/roshan30-git/picoclaw-scholar/pkg/memory"
	"github.com/roshan30-git/picoclaw-scholar/pkg/study"
	"github.com/roshan30-git/picoclaw-scholar/pkg/tools"
	"github.com/roshan30-git/picoclaw-scholar/pkg/visual"
)

type AgentLoop struct {
	cfg          *config.Config
	bus          *bus.MessageBus
	provider     tools.LLMProvider
	tools        map[string]tools.Tool
	mgr          *channels.Manager
	visParser    *visual.Parser
	router       *PersonaRouter
	calendar     *study.CalendarEngine
	reflections  *memory.ReflectionManager
	profileMgr   *memory.ProfileManager
	contextMgr   *memory.ContextManager
	smartHandler *study.SmartMessageHandler
	inbox        chan bus.InboundMessage
	quit         chan struct{}
	sessions     map[string][]tools.Message
	onShutdown   func()
}

func NewAgentLoop(cfg *config.Config, b *bus.MessageBus, provider tools.LLMProvider, vm *visual.Manager, router *PersonaRouter, cal *study.CalendarEngine, mem *memory.ReflectionManager, db *pkgdb.DB) *AgentLoop {
	return &AgentLoop{
		cfg:          cfg,
		bus:          b,
		provider:     provider,
		tools:        make(map[string]tools.Tool),
		visParser:    visual.NewParser(vm),
		router:       router,
		calendar:     cal,
		reflections:  mem,
		profileMgr:   memory.NewProfileManager(db.Conn()),
		contextMgr:   memory.NewContextManager(db, provider, memory.NewProfileManager(db.Conn()), study.NewDeadlineTracker(db)),
		smartHandler: study.NewSmartMessageHandler(provider, db),
		inbox:        b.Subscribe(),
		quit:         make(chan struct{}),
		sessions:     make(map[string][]tools.Message),
	}
}

func (l *AgentLoop) SetOnShutdown(f func()) {
	l.onShutdown = f
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

	if l.handleSmartPreprocessing(ctx, msg) {
		return
	}

	// 🛑 Manual Stop Command (Owner Only)
	if strings.TrimSpace(msg.Content) == "!stop" {
		owner := os.Getenv("STUDYCLAW_OWNER_NUMBER")
		if msg.From == owner || strings.Contains(msg.From, owner) {
			log.Printf("[AgentLoop] Shutdown command received from owner (%s)", msg.From)
			if l.mgr != nil {
				_ = l.mgr.Send(ctx, bus.OutboundMessage{
					ChatID: msg.ChatID, Content: "🛑 *StudyClaw is shutting down...* Goodbye!", Channel: msg.Channel,
				})
			}
			if l.onShutdown != nil {
				l.onShutdown()
			}
			return
		}
		log.Printf("[AgentLoop] Unauthorized !stop attempt from: %s", msg.From)
	}

	history := l.prepareEnrichedHistory(ctx, msg)

	resp, err := l.runAgentChat(ctx, history)
	if err != nil {
		log.Printf("[AgentLoop] Chat failed: %v", err)
		return
	}

	l.handlePostChatLogic(ctx, msg, resp, history)
}

func (l *AgentLoop) handleSmartPreprocessing(ctx context.Context, msg bus.InboundMessage) bool {
	if l.smartHandler == nil {
		return false
	}
	reply, continueToAgent := l.smartHandler.Process(ctx, msg.Content)
	if !continueToAgent {
		if reply != "" && l.mgr != nil {
			_ = l.mgr.Send(ctx, bus.OutboundMessage{
				ChatID: msg.ChatID, Content: reply, Channel: msg.Channel,
			})
		}
		return true
	}
	return false
}

func (l *AgentLoop) prepareEnrichedHistory(ctx context.Context, msg bus.InboundMessage) []tools.Message {
	history := l.getHistory(msg.Channel, msg.ChatID)
	l.detectCorrections(msg.Content, history)

	persona := l.router.RouteMessage(msg)
	summary := l.contextMgr.GetLatestChatSummary(msg.ChatID)
	enriched := l.enrichContext(msg.Content, persona)

	ctxBlock := l.contextMgr.BuildPrompt(ctx, msg.ChatID, msg.Content, summary)
	if ctxBlock != "" {
		enriched = ctxBlock + "\n\n" + enriched
	}

	return append(history, tools.Message{Role: "user", Content: enriched})
}

func (l *AgentLoop) runAgentChat(ctx context.Context, history []tools.Message) (*tools.LLMResponse, error) {
	toolDefs := l.getToolDefinitions()
	resp, err := l.provider.Chat(ctx, history, toolDefs, l.cfg.ModelName, nil)
	if err != nil {
		return nil, err
	}

	if len(resp.ToolCalls) > 0 {
		history = l.processToolCalls(ctx, history, resp.ToolCalls, toolDefs)
		return l.provider.Chat(ctx, history, toolDefs, l.cfg.ModelName, nil)
	}

	return resp, nil
}

func (l *AgentLoop) handlePostChatLogic(ctx context.Context, msg bus.InboundMessage, resp *tools.LLMResponse, history []tools.Message) {
	history = append(history, tools.Message{Role: "model", Content: resp.Content})

	if l.shouldTriggerSummary(history) {
		l.triggerRollingSummary(ctx, msg, resp.Content, history)
	} else {
		l.saveHistory(msg.Channel, msg.ChatID, history)
	}

	out := bus.OutboundMessage{
		ChatID:  msg.ChatID,
		Content: resp.Content,
		Channel: msg.Channel,
	}

	if l.visParser != nil {
		l.visParser.ApplyVisuals(&out)
	}

	if l.mgr != nil && out.Content != "" {
		_ = l.mgr.Send(ctx, out)
	}
}

func (l *AgentLoop) shouldTriggerSummary(history []tools.Message) bool {
	count := 0
	for _, m := range history {
		if m.Role == "user" {
			count++
		}
	}
	return count >= 3
}

func (l *AgentLoop) triggerRollingSummary(ctx context.Context, msg bus.InboundMessage, lastResp string, history []tools.Message) {
	l.contextMgr.SummarizeAndClear(ctx, msg.ChatID, history, func(newSummary string) {
		l.sessions[msg.Channel+":"+msg.ChatID] = []tools.Message{
			{Role: "model", Content: lastResp},
		}
	})
}

func (l *AgentLoop) getHistory(channel, chatID string) []tools.Message {
	return l.sessions[channel+":"+chatID]
}

func (l *AgentLoop) saveHistory(channel, chatID string, history []tools.Message) {
	if len(history) > 20 {
		history = history[len(history)-20:]
	}
	l.sessions[channel+":"+chatID] = history
}

func (l *AgentLoop) detectCorrections(content string, history []tools.Message) {
	if l.reflections == nil || len(history) == 0 {
		return
	}

	lowerMsg := strings.ToLower(content)
	keywords := []string{"wrong", "incorrect", "actually"}
	for _, kw := range keywords {
		if strings.Contains(lowerMsg, kw) {
			lastMsg := history[len(history)-1]
			if lastMsg.Role == "model" {
				log.Println("[AgentLoop] Correction detected! Logging reflection.")
				l.reflections.LogMistake(content, lastMsg.Content)
				return
			}
		}
	}
}

func (l *AgentLoop) enrichContext(content string, persona PersonaType) string {
	var preamble []string

	if personaPrompt := l.loadPersona(persona); personaPrompt != "" {
		preamble = append(preamble, personaPrompt)
	}

	if l.calendar != nil {
		preamble = append(preamble, l.calendar.GetContext())
	}

	if l.reflections != nil {
		if lessons := l.reflections.GetRecentReflections(); lessons != "" {
			preamble = append(preamble, lessons)
		}
	}

	if len(preamble) == 0 {
		return content
	}
	return strings.Join(preamble, "\n\n") + "\n\nUser Message:\n" + content
}

func (l *AgentLoop) loadPersona(persona PersonaType) string {
	promptName := string(persona)
	if persona == PersonaNone {
		promptName = "agent_explainer"
	}

	paths := []string{
		filepath.Join("workspace", "PROMPTS", "agents", promptName+".md"),
		filepath.Join("workspace", "PROMPTS", "agents", promptName+".txt"),
	}

	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			return fmt.Sprintf("🎭 SYSTEM PERSONA ACTIVE: %s\n%s", promptName, string(data))
		}
	}

	// Final fallback
	if base, err := os.ReadFile(filepath.Join("workspace", "PROMPTS", "agents", "base_soul.md")); err == nil {
		return "🎭 SYSTEM BASE SOUL:\n" + string(base)
	}
	return ""
}

func (l *AgentLoop) getToolDefinitions() []tools.ToolDefinition {
	var defs []tools.ToolDefinition
	for _, t := range l.tools {
		defs = append(defs, tools.ToolDefinition{
			Type: "function",
			Function: tools.ToolFunctionDefinition{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.Parameters(),
			},
		})
	}
	return defs
}

func (l *AgentLoop) processToolCalls(ctx context.Context, history []tools.Message, calls []tools.ToolCall, defs []tools.ToolDefinition) []tools.Message {
	history = append(history, tools.Message{Role: "model", Content: ""})
	for _, tc := range calls {
		log.Printf("Executing tool: %s", tc.Name)
		if t, ok := l.tools[tc.Name]; ok {
			res := t.Execute(ctx, tc.Args)
			resultMsg := fmt.Sprintf("Tool %s completed. Result: %s", tc.Name, res.ForLLM)
			history = append(history, tools.Message{Role: "user", Content: resultMsg})
		}
	}
	return history
}
