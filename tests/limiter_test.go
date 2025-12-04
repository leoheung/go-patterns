package tests

import (
	"testing"
	"time"
	"github.com/leoheung/go-patterns/parallel/limiter"
)

func TestStaticLimiter(t *testing.T) {
	limiter :=limiter.NewStaticLimiter(10 * time.Millisecond)
	start := time.Now()
	limiter.GrantNextToken()
	limiter.GrantNextToken()
	elapsed := time.Since(start)
	if elapsed < 20*time.Millisecond {
		t.Errorf("limiter did not wait enough, elapsed=%v", elapsed)
	}
	limiter.Stop()
}
