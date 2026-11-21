package rate

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

const key = "user:endpoint"

func TestLimiter_InvalidParams(t *testing.T) {
	tests := map[string]struct {
		burst float64
		rate  float64
	}{
		"zero burst": {0, 1},
		"zero rate":  {1, 0},
		"NaN burst":  {math.NaN(), 1},
		"NaN rate":   {1, math.NaN()},
		"Inf burst":  {math.Inf(1), 1},
		"Inf rate":   {1, math.Inf(1)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("NewLimiter(%v, %v) did not panic", tt.burst, tt.rate)
				}
			}()
			NewLimiter(t.Context(), tt.burst, tt.rate)
		})
	}
}

func TestLimiter_Burst(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(t.Context(), 10, 1)

		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				if ok, _ := l.Allow(key); !ok {
					t.Errorf("Allow(%q) = false, want true", key)
				}
			})
		}
		wg.Wait()

		if ok, _ := l.Allow(key); ok {
			t.Errorf("Allow(%q) after burst = true, want false", key)
		}
	})
}

func TestLimiter_KeysAreIndependent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(t.Context(), 1, 1)

		var wg sync.WaitGroup
		for i := range 100 {
			wg.Go(func() {
				k := fmt.Sprintf("user%d:endpoint", i)
				if ok, _ := l.Allow(k); !ok {
					t.Errorf("Allow(%q) = false, want true", k)
				}
				if ok, _ := l.Allow(k); ok {
					t.Errorf("second Allow(%q) = true, want false", k)
				}
			})
		}
		wg.Wait()
	})
}

func TestLimiter_Refill(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(t.Context(), 1, 1)

		if ok, _ := l.Allow(key); !ok {
			t.Fatalf("Allow(%q) = false, want true", key)
		}

		time.Sleep(500 * time.Millisecond)
		ok, wait := l.Allow(key)
		if ok {
			t.Errorf("Allow(%q) = true, want false", key)
		}
		if wait != (500 * time.Millisecond) {
			t.Errorf("Allow(%q).wait = %v, want 500ms", key, wait)
		}

		time.Sleep(500 * time.Millisecond)
		ok, wait = l.Allow(key)
		if !ok {
			t.Errorf("Allow(%q) = false, want true", key)
		}
		if wait != 0 {
			t.Errorf("Allow(%q).wait = %v, want 0", key, wait)
		}
	})
}

func TestLimiter_DeniedWaitIsPositive(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(t.Context(), 1, 1)

		l.mu.Lock()
		l.buckets[key] = &bucket{tokens: math.Nextafter(1, 0), last: time.Now()}
		l.mu.Unlock()

		ok, wait := l.Allow(key)
		if ok {
			t.Errorf("Allow(%q) = true, want false", key)
		}
		if wait <= 0 {
			t.Errorf("Allow(%q).wait = %v, want > 0", key, wait)
		}
	})
}

func TestLimiter_RefillCapsAtBurst(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(t.Context(), 4, 2)

		for range 4 {
			l.Allow(key)
		}
		time.Sleep(time.Hour)

		for range 4 {
			if ok, _ := l.Allow(key); !ok {
				t.Errorf("Allow(%q) = false, want true", key)
			}
		}
		if ok, _ := l.Allow(key); ok {
			t.Errorf("Allow(%q) beyond burst = true, want false", key)
		}
	})
}

func TestLimiter_Cleanup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Full after 10s, so cleanup runs every 10s.
		l := NewLimiter(t.Context(), 10, 1)
		time.Sleep(5 * time.Second)

		l.Allow("idle")
		for range 10 {
			l.Allow("busy")
		}

		// Step just past the first tick so cleanup has run: "idle" is full
		// again and dropped, "busy" has only 5 tokens and is kept.
		time.Sleep(5*time.Second + time.Nanosecond)

		l.mu.Lock()
		defer l.mu.Unlock()

		keys := slices.Collect(maps.Keys(l.buckets))
		if got, want := keys, []string{"busy"}; !slices.Equal(got, want) {
			t.Errorf("buckets.keys = %q, want %q", got, want)
		}
	})
}
