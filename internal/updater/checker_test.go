package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRemoteIsNewer(t *testing.T) {
	cases := []struct {
		remote, local string
		want          bool
	}{
		{"1.1.0", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"0.9.9", "1.0.0", false},
		{"2.0.0", "1.9.9", true},
		{"bad", "1.0.0", false},
		{"1.0.0", "bad", false},
	}
	for _, c := range cases {
		got := RemoteIsNewer(c.remote, c.local)
		if got != c.want {
			t.Errorf("RemoteIsNewer(%q, %q) = %v, want %v", c.remote, c.local, got, c.want)
		}
	}
}

// TestScheduleUpdateWith_IdleDoesNotSavePending verifies that the idle path
// does not persist a pending update file (it schedules the update instead).
func TestScheduleUpdateWith_IdleDoesNotSavePending(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled — goroutine will exit via ctx.Done() immediately

	p := NewPending(t.TempDir())
	ScheduleUpdateWith(ctx, func() bool { return true }, p, "1.1.0", "https://example.com/bin", func() {})

	_, ok := p.Load()
	if ok {
		t.Error("idle path should NOT save a pending update")
	}
}

// TestScheduleUpdateWith_IdleCallsRunFn verifies that runFn is eventually called
// when the system is idle. We override the 30 s delay by passing a runFn that
// signals immediately and a context that stays open long enough to receive it.
// Because the goroutine uses time.After(30s) we cannot wait 30 s in a unit
// test; instead we verify only that the goroutine is started and that ctx
// cancellation suppresses the call (the companion test above covers that).
// This test is intentionally skipped in short mode to keep CI fast.
func TestScheduleUpdateWith_IdleCallsRunFn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode: requires 30 s delay")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	p := NewPending(t.TempDir())
	called := make(chan struct{}, 1)

	ScheduleUpdateWith(ctx, func() bool { return true }, p, "1.1.0", "https://example.com/bin", func() {
		called <- struct{}{}
	})

	select {
	case <-called:
		// success
	case <-ctx.Done():
		t.Fatal("runFn was not called within timeout")
	}
}

func TestScheduleUpdateWith_NotIdleSavesPending(t *testing.T) {
	p := NewPending(t.TempDir())
	ScheduleUpdateWith(context.Background(), func() bool { return false }, p, "1.2.0", "https://example.com/bin", nil)

	u, ok := p.Load()
	if !ok {
		t.Fatal("expected pending update to be saved")
	}
	if u.Version != "1.2.0" {
		t.Errorf("version = %q, want %q", u.Version, "1.2.0")
	}
}

func TestCheckVersion_NoUpdateWhenCurrent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(versionResponse{Version: "0.0.1"}) // older than any real version
	}))
	defer srv.Close()

	orig := VersionServerURL
	VersionServerURL = srv.URL
	defer func() { VersionServerURL = orig }()

	n := &Notifier{}
	p := NewPending(t.TempDir())

	checkVersion(context.Background(), t.TempDir(), func() bool { return true }, n, p)

	_, ok := p.Load()
	if ok {
		t.Error("should not have saved pending update when version is current")
	}
}

func TestCheckVersion_SavesPendingWhenBusy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(versionResponse{
			Version:  "9.9.9",
			Binaries: map[string]string{binaryKey(): "https://example.com/bin"},
		})
	}))
	defer srv.Close()

	orig := VersionServerURL
	VersionServerURL = srv.URL
	defer func() { VersionServerURL = orig }()

	p := NewPending(t.TempDir())
	n := &Notifier{}

	checkVersion(context.Background(), t.TempDir(), func() bool { return false }, n, p)

	u, ok := p.Load()
	if !ok {
		t.Fatal("expected pending update when not idle")
	}
	if u.Version != "9.9.9" {
		t.Errorf("version = %q, want %q", u.Version, "9.9.9")
	}
}
