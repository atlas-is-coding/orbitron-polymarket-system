package config_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/config"
)

func TestConfigWatcher_NotifiesOnChange(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "config-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`[api]
clob_url = "https://clob.polymarket.com"
`)
	f.Close()

	reloaded := make(chan struct{}, 1)
	w, err := config.NewWatcher(f.Name(), func(_ *config.Config) {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go w.Run(ctx)

	time.Sleep(150 * time.Millisecond)

	err = os.WriteFile(f.Name(), []byte(`[api]
clob_url = "https://clob.polymarket.com"
timeout_sec = 15
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-reloaded:
		// pass
	case <-ctx.Done():
		t.Fatal("timeout: watcher did not fire")
	}
}

func TestConfigWatcher_NoFireOnUnchanged(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "config-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	content := `[api]
clob_url = "https://clob.polymarket.com"
`
	_, _ = f.WriteString(content)
	f.Close()

	fired := make(chan struct{}, 1)
	w, _ := config.NewWatcher(f.Name(), func(_ *config.Config) {
		select {
		case fired <- struct{}{}:
		default:
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()
	go w.Run(ctx)

	// don't touch the file
	select {
	case <-fired:
		t.Fatal("watcher fired without file change")
	case <-ctx.Done():
		// expected: no fire
	}
}
