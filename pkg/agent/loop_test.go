package agent

import (
	"context"
	"testing"
	"time"

	"github.com/roshan30-git/picoclaw-scholar/pkg/bus"
)

func TestAgentLoop_Run(t *testing.T) {
	t.Run("Context Cancellation", func(t *testing.T) {
		b := bus.NewMessageBus()
		loop := &AgentLoop{
			inbox: b.Subscribe(),
			quit:  make(chan struct{}),
		}

		ctx, cancel := context.WithCancel(context.Background())

		errCh := make(chan error)
		go func() {
			errCh <- loop.Run(ctx)
		}()

		cancel()

		select {
		case err := <-errCh:
			if err != context.Canceled {
				t.Errorf("expected context.Canceled, got %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for Run to exit")
		}
	})

	t.Run("Stop Channel", func(t *testing.T) {
		b := bus.NewMessageBus()
		loop := &AgentLoop{
			inbox: b.Subscribe(),
			quit:  make(chan struct{}),
		}

		ctx := context.Background()

		errCh := make(chan error)
		go func() {
			errCh <- loop.Run(ctx)
		}()

		loop.Stop()

		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("expected nil error on stop, got %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for Run to exit")
		}
	})

	t.Run("Handle Message - Manual Stop Command", func(t *testing.T) {
		// Mock STUDYCLAW_OWNER_NUMBER
		t.Setenv("STUDYCLAW_OWNER_NUMBER", "1234567890")

		b := bus.NewMessageBus()
		loop := &AgentLoop{
			inbox: b.Subscribe(),
			quit:  make(chan struct{}),
		}

		shutdownCalled := false
		shutdownDone := make(chan struct{})
		loop.SetOnShutdown(func() {
			shutdownCalled = true
			close(shutdownDone)
		})

		ctx := context.Background()

		go func() {
			_ = loop.Run(ctx)
		}()

		// Give Run loop time to start and subscribe to channels
		time.Sleep(10 * time.Millisecond)

		b.Publish(bus.InboundMessage{
			Content: "!stop",
			From:    "1234567890",
		})

		select {
		case <-shutdownDone:
			if !shutdownCalled {
				t.Error("expected onShutdown to be called")
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for !stop command to be processed")
		}

		// Cleanup
		loop.Stop()
	})
}
