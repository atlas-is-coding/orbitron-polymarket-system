package strategies

import "github.com/atlasdev/polytrade-bot/internal/copytrading"

// Executor places and closes orders on Polymarket.
// copytrading.OrderExecutor implements this interface.
type Executor interface {
	Open(assetID string, sizeUSD float64, negRisk bool) (*copytrading.OpenResult, error)
	Close(assetID string, sizeShares, avgBuyPrice float64, negRisk bool) (*copytrading.CloseResult, error)
	PlaceLimit(tokenID, side, orderType string, price, sizeUSD float64) (string, error)
}