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

	// Pre-process: smart handler decides how to treat the message
	if l.smartHandler != nil {
		reply, continueToAgent := l.smartHandler.Process(ctx, msg.Content)
		if !continueToAgent {
			if reply != "" && l.mgr != nil {
				_ = l.mgr.Send(ctx, bus.OutboundMessage{
					ChatID: msg.ChatID, Content: reply, Channel: msg.Channel,
				})
			}
			return
		}
	}

	history := l.getHistory(msg.Channel, msg.ChatID)
	l.detectCorrections(msg.Content, history)

	// Determine required persona for this turn
	persona := l.router.RouteMessage(msg)

	shortTermSummary := l.contextMgr.GetLatestChatSummary(msg.ChatID) // We'll add this wrapper method in loop soon, or just rely on ContextManager directly

	enrichedContent := l.enrichContext(msg.Content, persona)

	// Inject the dynamic Context Scratchpad built specifically for this message
	contextBlock := l.contextMgr.BuildPrompt(ctx, msg.ChatID, msg.Content, shortTermSummary)
	if contextBlock != "" {
		enrichedContent = contextBlock + "\n\n" + enrichedContent
	}

	history = append(history, tools.Message{Role: "user", Content: enrichedContent})

	toolDefs := l.getToolDefinitions()
	resp, err := l.provider.Chat(ctx, history, toolDefs, l.cfg.ModelName, nil)
	if err != nil {
		log.Printf("LLM error: %v", err)
		return
	}

	if len(resp.ToolCalls) > 0 {
		history = l.processToolCalls(ctx, history, resp.ToolCalls, toolDefs)
		resp, err = l.provider.Chat(ctx, history, toolDefs, l.cfg.ModelName, nil)
		if err != nil {
			log.Printf("Post-tool LLM error: %v", err)
			return
		}
	}

	history = append(history, tools.Message{Role: "model", Content: resp.Content})

	// Rolling Summary Logic: Every 3 logic pairs (6 messages: 3 user, 3 model), trigger summary
	var userMsgCount int
	for _, m := range history {
		if m.Role == "user" {
			userMsgCount++
		}
	}

	if userMsgCount >= 3 {
		l.contextMgr.SummarizeAndClear(ctx, msg.ChatID, history, func(newSummary string) {
			// Callback executes when generation is complete. Clear internal memory slice.
			// Keep ONLY the last model message so the immediate reply isn't fully lost mid-flight,
			// though the ContextManager will load the summary block next time anyway.
			l.sessions[msg.Channel+":"+msg.ChatID] = []tools.Message{
				{Role: "model", Content: resp.Content},
			}
		})
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
		if err := l.mgr.Send(ctx, out); err != nil {
			log.Printf("Failed to route response: %v", err)
		}
	}
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

	// 1. Inject Persona Overlays
	promptStr := string(persona)
	if persona == PersonaNone {
		promptStr = "agent_explainer" // safe default base
	}

	// Check .md first, then .txt
	personaPathMD := filepath.Join("workspace", "PROMPTS", "agents", promptStr+".md")
	personaPathTXT := filepath.Join("workspace", "PROMPTS", "agents", promptStr+".txt")

	var personaData []byte
	var err error
	if personaData, err = os.ReadFile(personaPathMD); err != nil {
		if personaData, err = os.ReadFile(personaPathTXT); err == nil {
			preamble = append(preamble, "🎭 SYSTEM PERSONA ACTIVE: "+promptStr+"\n"+string(personaData))
		} else {
			// Fallback to base_soul
			if baseData, err := os.ReadFile(filepath.Join("workspace", "PROMPTS", "agents", "base_soul.md")); err == nil {
				preamble = append(preamble, "🎭 SYSTEM BASE SOUL:\n"+string(baseData))
			}
		}
	} else {
		preamble = append(preamble, "🎭 SYSTEM PERSONA ACTIVE: "+promptStr+"\n"+string(personaData))
	}

	// 2. Inject Calendar
	if l.calendar != nil {
		preamble = append(preamble, l.calendar.GetContext())
	}
	// 3. Inject Autonomous Memory
	if l.reflections != nil {
		if lessons := l.reflections.GetRecentReflections(); lessons != "" {
			preamble = append(preamble, lessons)
		}
	}
	// 4. Student Learning Profile (personalization) - Moved entirely to ContextManager
	if len(preamble) == 0 {
		return content
	}
	return strings.Join(preamble, "\n\n") + "\n\nUser Message:\n" + content
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
