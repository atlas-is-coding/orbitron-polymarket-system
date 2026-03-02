package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/atlasdev/polytrade-bot/internal/config"
)

func TestMigrateAuthToWallet(t *testing.T) {
	toml := `
[api]
clob_url = "https://clob.polymarket.com"

[auth]
private_key = "deadbeef"
api_key     = "key123"
api_secret  = "secret123"
passphrase  = "pass123"
chain_id    = 137
`
	tmp := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmp, []byte(toml), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Wallets) != 1 {
		t.Fatalf("expected 1 wallet after migration, got %d", len(cfg.Wallets))
	}
	w := cfg.Wallets[0]
	if w.PrivateKey != "deadbeef" {
		t.Errorf("PrivateKey = %q, want deadbeef", w.PrivateKey)
	}
	if w.APIKey != "key123" {
		t.Errorf("APIKey = %q, want key123", w.APIKey)
	}
	if w.ChainID != 137 {
		t.Errorf("ChainID = %d, want 137", w.ChainID)
	}
	if w.Label != "Default" {
		t.Errorf("Label = %q, want Default", w.Label)
	}
	if w.ID == "" {
		t.Error("ID must not be empty after migration")
	}
	if !w.Enabled {
		t.Error("migrated wallet must be enabled")
	}
}

func TestLoadWallets(t *testing.T) {
	toml := `
[api]
clob_url = "https://clob.polymarket.com"

[[wallets]]
id          = "w1"
label       = "Main"
private_key = "aaa"
api_key     = "k1"
api_secret  = "s1"
passphrase  = "p1"
chain_id    = 137
enabled     = true
`
	tmp := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmp, []byte(toml), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Wallets) != 1 {
		t.Fatalf("expected 1 wallet, got %d", len(cfg.Wallets))
	}
	if cfg.Wallets[0].Label != "Main" {
		t.Errorf("Label = %q, want Main", cfg.Wallets[0].Label)
	}
}
