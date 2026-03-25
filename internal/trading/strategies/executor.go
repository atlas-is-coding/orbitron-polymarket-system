package strategies

import (
	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/copytrading"
)

// Executor places and closes orders on Polymarket.
// copytrading.OrderExecutor implements this interface.
type Executor interface {
	Open(assetID string, sizeUSD float64, negRisk bool) (*copytrading.OpenResult, error)
	Close(assetID string, sizeShares, avgBuyPrice float64, negRisk bool) (*copytrading.CloseResult, error)
	PlaceLimit(tokenID, side, orderType string, price, sizeUSD float64) (string, error)
}

// GammaClient defines the methods we use from gamma.Client.
type GammaClient interface {
	GetMarkets(params gamma.MarketsParams) ([]gamma.Market, error)
	GetMarket(id string) (*gamma.Market, error)
}

// OrderTagger records which strategy placed a given order.
// analytics.AnalyticsHub implements this interface.
type OrderTagger interface {
	TagOrder(orderID, strategy string)
}

// TaggingExecutor wraps an Executor and tags every successfully placed order
// with the strategy name, so the analytics system can resolve it later.
type TaggingExecutor struct {
	inner    Executor
	tagger   OrderTagger
	strategy string
}

// NewTaggingExecutor returns an Executor that tags orders via tagger.
// If tagger is nil the wrapper is a transparent pass-through.
func NewTaggingExecutor(inner Executor, tagger OrderTagger, strategyName string) Executor {
	if inner == nil {
		return nil
	}
	return &TaggingExecutor{inner: inner, tagger: tagger, strategy: strategyName}
}

func (e *TaggingExecutor) Open(assetID string, sizeUSD float64, negRisk bool) (*copytrading.OpenResult, error) {
	res, err := e.inner.Open(assetID, sizeUSD, negRisk)
	if err == nil && e.tagger != nil && res != nil && res.OrderID != "" {
		e.tagger.TagOrder(res.OrderID, e.strategy)
	}
	return res, err
}

func (e *TaggingExecutor) Close(assetID string, sizeShares, avgBuyPrice float64, negRisk bool) (*copytrading.CloseResult, error) {
	return e.inner.Close(assetID, sizeShares, avgBuyPrice, negRisk)
}

func (e *TaggingExecutor) PlaceLimit(tokenID, side, orderType string, price, sizeUSD float64) (string, error) {
	orderID, err := e.inner.PlaceLimit(tokenID, side, orderType, price, sizeUSD)
	if err == nil && e.tagger != nil && orderID != "" {
		e.tagger.TagOrder(orderID, e.strategy)
	}
	return orderID, err
}
