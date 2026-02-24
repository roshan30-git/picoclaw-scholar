package bus

import "sync"

type InboundMessage struct {
	From     string
	ChatID   string
	Content  string
	Media    []byte
	Metadata map[string]string
	Channel  string
}

type OutboundMessage struct {
	ChatID  string
	Content string
	Channel string
}

type MessageBus struct {
	mu   sync.RWMutex
	subs []chan InboundMessage
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
