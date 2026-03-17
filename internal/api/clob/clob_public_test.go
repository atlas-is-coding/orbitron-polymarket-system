package clob_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/orbitron/internal/testutil"
)

func TestGetMarkets_FirstPage(t *testing.T) {
	client := testutil.NewCLOBClient()
	resp, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Data)
	assert.NotEmpty(t, resp.Data[0].ConditionID)
}

func TestGetMarkets_Pagination(t *testing.T) {
	client := testutil.NewCLOBClient()
	page1, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page1.NextCursor)
	page2, err := client.GetMarkets(page1.NextCursor)
	require.NoError(t, err)
	require.NotNil(t, page2)
	assert.NotEqual(t, page1.Data[0].ConditionID, page2.Data[0].ConditionID)
}

func TestGetMarket_ByConditionID(t *testing.T) {
	client := testutil.NewCLOBClient()
	page, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page.Data)
	condID := page.Data[0].ConditionID
	m, err := client.GetMarket(condID)
	require.NoError(t, err)
	assert.Equal(t, condID, m.ConditionID)
}

func getFirstTokenID(t *testing.T) string {
	t.Helper()
	client := testutil.NewCLOBClient()
	page, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page.Data)

	// Try first 10 markets to find one that has an orderbook
	for i := 0; i < len(page.Data) && i < 10; i++ {
		m := page.Data[i]
		if len(m.Tokens) == 0 {
			continue
		}
		tokenID := m.Tokens[0].TokenID
		_, err := client.GetOrderBook(tokenID)
		if err == nil {
			return tokenID
		}
	}

	return page.Data[0].Tokens[0].TokenID
}

func TestGetOrderBook_ReturnsBook(t *testing.T) {
	tokenID := getFirstTokenID(t)
	client := testutil.NewCLOBClient()
	ob, err := client.GetOrderBook(tokenID)
	require.NoError(t, err)
	assert.Equal(t, tokenID, ob.AssetID)
	assert.NotNil(t, ob.Bids)
	assert.NotNil(t, ob.Asks)
}

func TestGetMidpoint(t *testing.T) {
	tokenID := getFirstTokenID(t)
	client := testutil.NewCLOBClient()
	mid, err := client.GetMidpoint(tokenID)
	require.NoError(t, err)
	assert.NotEmpty(t, mid.Mid)
}

func TestGetPrice_BuyAndSell(t *testing.T) {
	tokenID := getFirstTokenID(t)
	client := testutil.NewCLOBClient()
	buy, err := client.GetPrice(tokenID, "BUY")
	require.NoError(t, err)
	assert.NotEmpty(t, buy.Price)
	sell, err := client.GetPrice(tokenID, "SELL")
	require.NoError(t, err)
	assert.NotEmpty(t, sell.Price)
}

func TestGetSpread(t *testing.T) {
	tokenID := getFirstTokenID(t)
	client := testutil.NewCLOBClient()
	spread, err := client.GetSpread(tokenID)
	require.NoError(t, err)
	assert.NotEmpty(t, spread.Spread)
}

func TestGetMarketTrades_PublicEndpoint(t *testing.T) {
	tokenID := getFirstTokenID(t)
	client := testutil.NewCLOBClient()
	trades, err := client.GetMarketTrades(tokenID, 5)
	require.NoError(t, err)
	assert.NotNil(t, trades)
}
