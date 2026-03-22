package builder_test

import (
	"sync"
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/builder"
	"github.com/rs/zerolog"
)

func newTestLogger() *builder.OrderExecutionLogger {
	return builder.NewOrderExecutionLogger(zerolog.Nop())
}

func TestLogger_CountsWithKey(t *testing.T) {
	l := newTestLogger()
	l.LogOrder(builder.OrderLogEntry{OrderID: "a", BuilderKeySet: true, Timestamp: time.Now(), Success: true})
	l.LogOrder(builder.OrderLogEntry{OrderID: "b", BuilderKeySet: true, Timestamp: time.Now(), Success: true})

	total, withKey, withoutKey := l.Summary()
	if total != 2 {
		t.Fatalf("total: want 2, got %d", total)
	}
	if withKey != 2 {
		t.Fatalf("withKey: want 2, got %d", withKey)
	}
	if withoutKey != 0 {
		t.Fatalf("withoutKey: want 0, got %d", withoutKey)
	}
}

func TestLogger_CountsWithoutKey(t *testing.T) {
	l := newTestLogger()
	l.LogOrder(builder.OrderLogEntry{OrderID: "a", BuilderKeySet: false, Timestamp: time.Now(), Success: true})

	total, withKey, withoutKey := l.Summary()
	if total != 1 {
		t.Fatalf("total: want 1, got %d", total)
	}
	if withKey != 0 {
		t.Fatalf("withKey: want 0, got %d", withKey)
	}
	if withoutKey != 1 {
		t.Fatalf("withoutKey: want 1, got %d", withoutKey)
	}
}

func TestLogger_ThreadSafety(t *testing.T) {
	l := newTestLogger()
	const n = 500
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.LogOrder(builder.OrderLogEntry{BuilderKeySet: true, Timestamp: time.Now(), Success: true})
		}()
	}
	wg.Wait()

	total, withKey, _ := l.Summary()
	if total != n {
		t.Fatalf("total: want %d, got %d", n, total)
	}
	if withKey != n {
		t.Fatalf("withKey: want %d, got %d", n, withKey)
	}
}

func TestLogger_SummaryEvery100(t *testing.T) {
	// Just verify no panic at 100th order boundary.
	l := newTestLogger()
	for i := 0; i < 100; i++ {
		l.LogOrder(builder.OrderLogEntry{BuilderKeySet: true, Timestamp: time.Now(), Success: true})
	}
	total, _, _ := l.Summary()
	if total != 100 {
		t.Fatalf("want 100, got %d", total)
	}
}
