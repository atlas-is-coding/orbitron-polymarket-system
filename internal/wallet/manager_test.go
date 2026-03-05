package wallet_test

import (
	"testing"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/wallet"
)

func TestManagerAddGet(t *testing.T) {
	mgr := wallet.NewManager(nil)
	cfg := config.WalletConfig{
		ID:      "test-id",
		Label:   "Test",
		Enabled: true,
	}
	mgr.AddInactive(cfg)

	inst, ok := mgr.Get("test-id")
	if !ok {
		t.Fatal("Get: wallet not found")
	}
	if inst.Cfg.Label != "Test" {
		t.Errorf("Label = %q, want Test", inst.Cfg.Label)
	}
	if inst.Stats == nil {
		t.Error("Stats must not be nil")
	}
}

func TestManagerWallets(t *testing.T) {
	mgr := wallet.NewManager(nil)
	mgr.AddInactive(config.WalletConfig{ID: "w1", Label: "One", Enabled: true})
	mgr.AddInactive(config.WalletConfig{ID: "w2", Label: "Two", Enabled: false})

	all := mgr.Wallets()
	if len(all) != 2 {
		t.Fatalf("expected 2 wallets, got %d", len(all))
	}
}

func TestManagerUpdateLabel(t *testing.T) {
	mgr := wallet.NewManager(nil)
	mgr.AddInactive(config.WalletConfig{ID: "w1", Label: "Old", Enabled: true})

	if err := mgr.UpdateLabel("w1", "New"); err != nil {
		t.Fatal(err)
	}
	inst, _ := mgr.Get("w1")
	if inst.Cfg.Label != "New" {
		t.Errorf("Label = %q, want New", inst.Cfg.Label)
	}
}

func TestManagerUpdateLabelNotFound(t *testing.T) {
	mgr := wallet.NewManager(nil)
	err := mgr.UpdateLabel("nonexistent", "X")
	if err == nil {
		t.Error("expected error for non-existent wallet, got nil")
	}
}

func TestManagerRemove(t *testing.T) {
	mgr := wallet.NewManager(nil)
	mgr.AddInactive(config.WalletConfig{ID: "w1", Label: "One", Enabled: true})
	mgr.AddInactive(config.WalletConfig{ID: "w2", Label: "Two", Enabled: true})

	if err := mgr.Remove("w1"); err != nil {
		t.Fatal(err)
	}
	all := mgr.Wallets()
	if len(all) != 1 {
		t.Fatalf("expected 1 wallet after remove, got %d", len(all))
	}
	if all[0].Cfg.ID != "w2" {
		t.Errorf("remaining wallet ID = %q, want w2", all[0].Cfg.ID)
	}
}

func TestManagerToggle(t *testing.T) {
	mgr := wallet.NewManager(nil)
	mgr.AddInactive(config.WalletConfig{ID: "w1", Label: "One", Enabled: true})

	if err := mgr.Toggle("w1", false); err != nil {
		t.Fatal(err)
	}
	inst, _ := mgr.Get("w1")
	if inst.Cfg.Enabled {
		t.Error("expected Enabled=false after Toggle(false)")
	}

	if err := mgr.Toggle("w1", true); err != nil {
		t.Fatal(err)
	}
	inst, _ = mgr.Get("w1")
	if !inst.Cfg.Enabled {
		t.Error("expected Enabled=true after Toggle(true)")
	}
}

func TestManagerPrimary(t *testing.T) {
	mgr := wallet.NewManager(nil)
	mgr.AddInactive(config.WalletConfig{ID: "w1", Label: "One", Enabled: true, Primary: false})
	mgr.AddInactive(config.WalletConfig{ID: "w2", Label: "Two", Enabled: true, Primary: true})

	p := mgr.Primary()
	if p == nil {
		t.Fatal("Primary: expected wallet, got nil")
	}
	if p.Cfg.ID != "w2" {
		t.Fatalf("Primary: expected w2, got %s", p.Cfg.ID)
	}
}

func TestManagerPrimaryFallback(t *testing.T) {
	mgr := wallet.NewManager(nil)
	mgr.AddInactive(config.WalletConfig{ID: "w1", Label: "One", Enabled: true, Primary: false})

	p := mgr.Primary()
	if p == nil {
		t.Fatal("Primary fallback: expected first enabled wallet, got nil")
	}
	if p.Cfg.ID != "w1" {
		t.Fatalf("Primary fallback: expected w1, got %s", p.Cfg.ID)
	}
}

func TestManagerSetPrimary(t *testing.T) {
	mgr := wallet.NewManager(nil)
	mgr.AddInactive(config.WalletConfig{ID: "w1", Label: "One", Enabled: true, Primary: true})
	mgr.AddInactive(config.WalletConfig{ID: "w2", Label: "Two", Enabled: true, Primary: false})

	if err := mgr.SetPrimary("w2"); err != nil {
		t.Fatalf("SetPrimary: %v", err)
	}
	p := mgr.Primary()
	if p.Cfg.ID != "w2" {
		t.Fatalf("SetPrimary: expected w2 primary, got %s", p.Cfg.ID)
	}
	w1, _ := mgr.Get("w1")
	if w1.Cfg.Primary {
		t.Fatal("SetPrimary: w1 should no longer be primary")
	}
}

func TestManagerSetPrimaryNotFound(t *testing.T) {
	mgr := wallet.NewManager(nil)
	mgr.AddInactive(config.WalletConfig{ID: "w1", Label: "One", Enabled: true, Primary: true})

	err := mgr.SetPrimary("nonexistent")
	if err == nil {
		t.Fatal("SetPrimary nonexistent: expected error, got nil")
	}
	// w1 must still be primary (state not corrupted)
	p := mgr.Primary()
	if p == nil || p.Cfg.ID != "w1" {
		t.Fatalf("SetPrimary nonexistent: w1 should still be primary, got %v", p)
	}
}

func TestWalletStats(t *testing.T) {
	stats := &wallet.WalletStats{}
	stats.Set(100.0, 10.0, 3, 15)
	bal, pnl, orders, total := stats.Get()
	if bal != 100.0 {
		t.Errorf("BalanceUSD = %f, want 100.0", bal)
	}
	if pnl != 10.0 {
		t.Errorf("PnLUSD = %f, want 10.0", pnl)
	}
	if orders != 3 {
		t.Errorf("OpenOrders = %d, want 3", orders)
	}
	if total != 15 {
		t.Errorf("TotalTrades = %d, want 15", total)
	}
}
