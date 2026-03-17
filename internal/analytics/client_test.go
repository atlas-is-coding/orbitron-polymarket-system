package analytics

import (
	"testing"
	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/ethereum/go-ethereum/crypto"
	"encoding/hex"
)

func TestClient_SignReport(t *testing.T) {
	pk, _ := crypto.GenerateKey()
	pkHex := hex.EncodeToString(crypto.FromECDSA(pk))
	signer, _ := auth.NewL1Signer(pkHex)

	client := &Client{
		signer:  signer,
		address: signer.Address(),
		label:   "Test Bot",
	}

	trades := []TradeReport{
		{ID: "t1", Volume: 100},
	}

	payload, signature, err := client.preparePayload(trades)
	assert.NoError(t, err)

	assert.Equal(t, client.address, payload.Address)
	assert.Equal(t, "Test Bot", payload.Label)
	assert.NotEmpty(t, signature)
}
