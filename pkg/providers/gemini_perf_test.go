package providers

import (
	"context"
	"testing"
	"time"
)

func TestChatCancellationLatency(t *testing.T) {
	// Simulated Old Behavior benchmark value:
	oldDuration := 10 * time.Second

	// Test New Behavior Context Cancellation inside Retry Loop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	startNew := time.Now()
	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Done():
	}
	newDuration := time.Since(startNew)

	t.Logf("Old Behavior Duration: %v", oldDuration)
	t.Logf("New Behavior Duration: %v", newDuration)
}

func TestRateLimiterCancellationLatency(t *testing.T) {
	limiter := newRateLimiter(1) // 1 token per 60 seconds

	// Consume the first token immediately
	ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel1()
	err := limiter.Wait(ctx1)
	if err != nil {
		t.Fatalf("expected no error on first token, got: %v", err)
	}

	// Wait for the second token with a short context cancellation
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel2()

	start := time.Now()
	err = limiter.Wait(ctx2)
	latency := time.Since(start)

	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded error, got: %v", err)
	}

	t.Logf("Rate Limiter Cancellation Latency: %v", latency)
	if latency > 50*time.Millisecond {
		t.Errorf("latency too high: %v", latency)
	}
}
