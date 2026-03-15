package strategies_test

import (
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/trading/strategies"
)

func TestFadeDetectsSpike(t *testing.T) {
	// prev=0.40, curr=0.52 → rise = 30% → above 10% threshold
	found, risePct := strategies.DetectSpike(0.40, 0.52, 10.0)
	if !found {
		t.Fatalf("expected spike detected: prev=0.40 curr=0.52")
	}
	if risePct < 29 || risePct > 31 {
		t.Fatalf("expected ~30%% rise, got %.2f", risePct)
	}
}

func TestFadeNoSpikeSmallMove(t *testing.T) {
	// prev=0.50, curr=0.53 → rise = 6% → below 10% threshold
	found, _ := strategies.DetectSpike(0.50, 0.53, 10.0)
	if found {
		t.Fatal("expected no spike: rise too small")
	}
}

func TestFadeNegativeMove(t *testing.T) {
	// Price went DOWN — no spike
	found, _ := strategies.DetectSpike(0.60, 0.50, 10.0)
	if found {
		t.Fatal("expected no spike on price decrease")
	}
}

func TestFadeCooldownPreventsRepeat(t *testing.T) {
	tracker := strategies.NewCooldownTracker(300)
	tracker.Record("cid1")
	if !tracker.InCooldown("cid1") {
		t.Fatal("expected cid1 to be in cooldown after record")
	}
	if tracker.InCooldown("cid2") {
		t.Fatal("expected cid2 not in cooldown")
	}
}

func TestFadeCooldownExpires(t *testing.T) {
	tracker := strategies.NewCooldownTracker(0) // 0 seconds = instant expire
	tracker.Record("cid1")
	time.Sleep(10 * time.Millisecond)
	if tracker.InCooldown("cid1") {
		t.Fatal("expected cooldown to expire")
	}
}
