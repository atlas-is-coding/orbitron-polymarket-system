package config

import (
	"crypto/rand"
	"encoding/hex"
)

// migrateAuth converts legacy [auth] section into first [[wallets]] entry.
// Called by Load() when Wallets is empty but Auth has credentials.
func (c *Config) migrateAuth() {
	if len(c.Wallets) > 0 {
		return
	}
	a := c.Auth
	if a.PrivateKey == "" && a.APIKey == "" {
		return
	}
	chainID := a.ChainID
	if chainID == 0 {
		chainID = 137
	}
	negRisk := c.Trading.NegRisk
	c.Wallets = []WalletConfig{
		{
			ID:         newWalletID(),
			Label:      "Default",
			PrivateKey: a.PrivateKey,
			APIKey:     a.APIKey,
			APISecret:  a.APISecret,
			Passphrase: a.Passphrase,
			ChainID:    chainID,
			Enabled:    true,
			NegRisk:    negRisk,
		},
	}
}

// newWalletID generates a short random hex ID for a wallet.
func newWalletID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
