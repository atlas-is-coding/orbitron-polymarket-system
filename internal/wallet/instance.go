package wallet

import (
	"context"

	"github.com/atlasdev/polytrade-bot/internal/config"
)

// WalletInstance holds the runtime state of one wallet.
// Subsystem fields (TradesMon, CopyTrader) will be populated in Task 4
// when the wallet is started by WalletManager.AddActive.
type WalletInstance struct {
	Cfg     config.WalletConfig
	Address string // derived from private_key via L1Signer; empty if no private_key
	Stats   *WalletStats
	cancel  context.CancelFunc
}

// Stop cancels the wallet's context, triggering graceful shutdown of its subsystems.
func (w *WalletInstance) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}
