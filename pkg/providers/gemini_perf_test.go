package providers

import (
	"context"
	"errors"
	"testing"
)

// BenchmarkRateLimiter_Wait_Cancellation measures the latency of Wait when the context is cancelled.
// We expect this to drop from ~500ms (due to sleep) down to nanoseconds.
func BenchmarkRateLimiter_Wait_Cancellation(b *testing.B) {
	// Create a rate limiter with 0 tokens and a very slow refill rate.
	limiter := newRateLimiter(1) // 1 request per minute (rpm)
	// drain the single token immediately
	limiter.tokens = 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())

		// Run Wait in a goroutine so we can cancel it
		done := make(chan error)
		go func() {
			done <- limiter.Wait(ctx)
		}()

		// cancel immediately
		cancel()

		// wait for Wait to return
		err := <-done
		if !errors.Is(err, context.Canceled) {
			b.Fatalf("expected context cancellation error, got: %v", err)
		}
	}
}
