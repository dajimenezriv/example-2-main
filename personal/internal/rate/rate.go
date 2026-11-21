package rate

import (
	"context"
	"math"
	"sync"
	"time"
)

type Limiter struct {
	burst float64
	rate  float64 // tokens per second

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewLimiter(ctx context.Context, burst float64, rate float64) *Limiter {
	if !(burst >= 1) || !(rate > 0) || math.IsInf(burst, 0) || math.IsInf(rate, 0) {
		panic("rate: invalid burst or rate")
	}

	l := &Limiter{
		burst:   burst,
		rate:    rate,
		buckets: make(map[string]*bucket),
	}
	go l.cleanupLoop(ctx)
	return l
}

// Allow tries to take a token from the bucket. If the key is new it creates a new
// bucket filled with tokens.
// If the bucket has tokens it takes one and returns true.
// If the bucket is empty it returns false and how long until the next token, which is
// always positive.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}

	l.refill(b, now)

	if b.tokens < 1 {
		// Round up so a wait under 1ns doesn't truncate to 0.
		// (1 - 0.9999999999) tokens / 1 token/s
		// = 0.0000000001 s
		// = 0.0000001 ms
		// = 0.0001 µs
		// = 0.1 ns, which time.Duration would truncate to 0.
		wait := time.Duration(math.Ceil((1 - b.tokens) / l.rate * float64(time.Second)))
		return false, wait
	}

	b.tokens--
	return true, 0
}

func (l *Limiter) refill(b *bucket, now time.Time) {
	added := l.rate * now.Sub(b.last).Seconds()
	b.tokens = min(b.tokens+added, l.burst)
	b.last = now
}

func (l *Limiter) cleanupLoop(ctx context.Context) {
	t := l.burst / l.rate
	ticker := time.NewTicker(time.Duration(t * float64(time.Second)))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.cleanup()
		}
	}
}

func (l *Limiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for key, b := range l.buckets {
		l.refill(b, now)
		if b.tokens >= l.burst {
			delete(l.buckets, key)
		}
	}
}
