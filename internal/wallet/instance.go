package wallet

import (
	"context"

	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/auth"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/copytrading"
	"github.com/atlasdev/polytrade-bot/internal/monitor"
)

// WalletInstance holds the runtime state of one wallet.
// Subsystem fields (TradesMon, CopyTrader) are populated in main.go
// when the wallet is started by WalletManager.AddActive.
type WalletInstance struct {
	Cfg        config.WalletConfig
	Address    string               // derived from private_key via L1Signer; empty if no private_key
	L2         *auth.L2Credentials  // nil for watch-only wallets
	ClobClient *clob.Client         // nil for watch-only wallets
	TradesMon  *monitor.TradesMonitor  // nil if trades monitoring disabled
	CopyTrader *copytrading.CopyTrader // nil if copytrading disabled
	Stats      *WalletStats
	cancel     context.CancelFunc
}

// Stop cancels the wallet's context, triggering graceful shutdown of its subsystems.
func (w *WalletInstance) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}
