package channels

import (
	"context"
	"log"
	"sync"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
)

type Channel interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Send(ctx context.Context, msg bus.OutboundMessage) error
	IsRunning() bool
}

type Manager struct {
	mu       sync.RWMutex
	channels map[string]Channel
}

func NewManager() *Manager {
	return &Manager{channels: make(map[string]Channel)}
}

func (m *Manager) Register(ch Channel) {
	m.mu.Lock()
	m.channels[ch.Name()] = ch
	m.mu.Unlock()
}

func (m *Manager) StartAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, ch := range m.channels {
		if err := ch.Start(ctx); err != nil {
			return err
		}
		log.Printf("Channel started: %s", ch.Name())
	}
	return nil
}

func (m *Manager) StopAll(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, ch := range m.channels {
		ch.Stop(ctx)
	}
}

func (m *Manager) Send(ctx context.Context, msg bus.OutboundMessage) error {
	m.mu.RLock()
	ch, ok := m.channels[msg.Channel]
	m.mu.RUnlock()
	
	if ok {
		return ch.Send(ctx, msg)
	}
	return fmt.Errorf("channel %s not found", msg.Channel)
}

// BaseChannel provides common functionality for channel implementations.
type BaseChannel struct {
	name    string
	running bool
	mu      sync.RWMutex
	bus     *bus.MessageBus
	allowed []string
}

func NewBaseChannel(name string, _ any, b *bus.MessageBus, allowed []string) *BaseChannel {
	return &BaseChannel{name: name, bus: b, allowed: allowed}
}

func (b *BaseChannel) Name() string { return b.name }

func (b *BaseChannel) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

func (b *BaseChannel) SetRunning(v bool) {
	b.mu.Lock()
	b.running = v
	b.mu.Unlock()
}

func (b *BaseChannel) HandleMessage(from, chatID, text string, media []byte, meta map[string]string) {
	msg := bus.InboundMessage{
		From:     from,
		ChatID:   chatID,
		Content:  text,
		Media:    media,
		Metadata: meta,
		Channel:  b.name,
	}
	b.bus.Publish(msg)
}
