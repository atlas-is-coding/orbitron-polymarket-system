//go:build integration

package clob_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/testutil"
)

func TestDeriveAPIKey_FromPrivateKey(t *testing.T) {
	l1 := testutil.LoadL1Signer(t)
	client := testutil.NewCLOBClient()
	creds, err := client.DeriveAPIKey(l1)
	require.NoError(t, err)
	assert.NotEmpty(t, creds.APIKey)
	assert.NotEmpty(t, creds.APISecret)
	assert.NotEmpty(t, creds.Passphrase)
	assert.Equal(t, l1.Address(), creds.Address)
}

func TestGetOrders_Authenticated(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)
	resp, err := client.GetOrders()
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Data)
}

func TestGetPositions_Authenticated(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)
	positions, err := client.GetPositions()
	require.NoError(t, err)
	assert.NotNil(t, positions)
}

func TestGetTrades_Authenticated(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)
	resp, err := client.GetTrades()
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Data)
}

func TestGetBalanceAllowance_USDC(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)
	bal, err := client.GetBalanceAllowance("COLLATERAL", "")
	require.NoError(t, err)
	assert.NotNil(t, bal)
	assert.Equal(t, "COLLATERAL", bal.AssetType)
	assert.NotEmpty(t, bal.Balance)
}

func TestGetDataOrders_Authenticated(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)
	resp, err := client.GetDataOrders(clob.OrdersFilter{})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}
