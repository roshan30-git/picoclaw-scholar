package bus

import "sync"

type InboundMessage struct {
	From     string
	ChatID   string
	Content  string
	Media    []string
	Metadata map[string]string
	Channel  string
}

type OutboundMessage struct {
	ChatID     string
	Content    string
	Channel    string
	VisualID   string
	VisualType string
}

type MessageBus struct {
	mu      sync.RWMutex
	subs    []chan InboundMessage
	outSubs []chan OutboundMessage
}

func NewMessageBus() *MessageBus {
	return &MessageBus{}
}

func (b *MessageBus) Subscribe() chan InboundMessage {
	ch := make(chan InboundMessage, 100)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	return ch
}

func (b *MessageBus) Publish(msg InboundMessage) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (b *MessageBus) SubscribeOutbound() chan OutboundMessage {
	ch := make(chan OutboundMessage, 100)
	b.mu.Lock()
	b.outSubs = append(b.outSubs, ch)
	b.mu.Unlock()
	return ch
}

func (b *MessageBus) PublishOutbound(msg OutboundMessage) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.outSubs {
		select {
		case ch <- msg:
		default:
		}
	}
}
