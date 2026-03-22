package copytrading

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"time"

	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/atlasdev/orbitron/internal/builder"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

const (
	// usdcDecimals — количество десятичных знаков USDC и условных токенов (6)
	usdcDecimals = 1_000_000
	// defaultFeeRateBps — fee по умолчанию (0 = автоматически)
	defaultFeeRateBps = 0
)

// OrderExecutor размещает и закрывает ордера на Polymarket через CLOB API.
type OrderExecutor struct {
	clob          *clob.Client
	orderSigner   *auth.OrderSigner
	apiKey        string
	makerAddr     string
	logger        zerolog.Logger
	builderAPIKey string                        // опциональный ключ Builder Program
	orderLogger   *builder.OrderExecutionLogger // nil if not configured
}

// WithBuilderKey устанавливает Polymarket Builder API key.
// Возвращает executor для chain-вызовов.
func (e *OrderExecutor) WithBuilderKey(key string) *OrderExecutor {
	e.builderAPIKey = key
	return e
}

// WithOrderLogger attaches an OrderExecutionLogger for attribution auditing.
func (e *OrderExecutor) WithOrderLogger(l *builder.OrderExecutionLogger) *OrderExecutor {
	e.orderLogger = l
	return e
}

// logOrder records order attribution if a logger is configured.
func (e *OrderExecutor) logOrder(orderID string, success bool) {
	if e.orderLogger == nil {
		return
	}
	e.orderLogger.LogOrder(builder.OrderLogEntry{
		OrderID:       orderID,
		BuilderKeySet: e.builderAPIKey != "",
		Timestamp:     time.Now(),
		Success:       success,
	})
}

// NewOrderExecutor создаёт OrderExecutor.
// apiKey — значение поля "owner" в теле запроса (L2 api_key).
func NewOrderExecutor(
	clobClient *clob.Client,
	orderSigner *auth.OrderSigner,
	apiKey string,
	makerAddr string,
	log zerolog.Logger,
) *OrderExecutor {
	return &OrderExecutor{
		clob:        clobClient,
		orderSigner: orderSigner,
		apiKey:      apiKey,
		makerAddr:   makerAddr,
		logger:      log.With().Str("component", "order-executor").Logger(),
	}
}

// OpenResult — результат успешного открытия позиции.
type OpenResult struct {
	OrderID string
	Price   float64
	Size    float64
}

// CloseResult — результат успешного закрытия позиции.
type CloseResult struct {
	OrderID string
	Price   float64
	PnL     float64
}

// Open размещает market-buy ордер для указанного токена.
// assetID — token_id (ERC-1155), sizeUSD — размер позиции в USD.
// negRisk — true для рынков с несколькими взаимоисключающими исходами.
func (e *OrderExecutor) Open(assetID string, sizeUSD float64, negRisk bool) (*OpenResult, error) {
	priceResp, err := e.clob.GetPrice(assetID, "BUY")
	if err != nil {
		return nil, fmt.Errorf("executor: get BUY price for %s: %w", assetID, err)
	}
	price, err := strconv.ParseFloat(priceResp.Price, 64)
	if err != nil || price <= 0 {
		return nil, fmt.Errorf("executor: invalid BUY price %q for %s", priceResp.Price, assetID)
	}

	// Размер в токенах: сколько токенов купить за sizeUSD
	sizeShares := sizeUSD / price

	req, err := e.buildOrderRequest(assetID, price, sizeShares, auth.Buy)
	if err != nil {
		return nil, fmt.Errorf("executor: build BUY order: %w", err)
	}

	resp, err := e.clob.CreateOrder(req)
	if err != nil {
		return nil, fmt.Errorf("executor: CreateOrder BUY: %w", err)
	}
	if !resp.Success {
		e.logOrder(resp.OrderID, false)
		return nil, fmt.Errorf("executor: BUY order rejected: %s", resp.ErrorMsg)
	}

	e.logOrder(resp.OrderID, true)
	e.logger.Info().
		Str("asset_id", assetID).
		Str("order_id", resp.OrderID).
		Float64("price", price).
		Float64("size_shares", sizeShares).
		Float64("size_usd", sizeUSD).
		Msg("opened copy position")

	return &OpenResult{OrderID: resp.OrderID, Price: price, Size: sizeShares}, nil
}

// Close продаёт позицию по лучшей цене bid.
// sizeShares — количество токенов для продажи.
// avgBuyPrice — средняя цена покупки (для расчёта P&L).
func (e *OrderExecutor) Close(assetID string, sizeShares, avgBuyPrice float64, negRisk bool) (*CloseResult, error) {
	priceResp, err := e.clob.GetPrice(assetID, "SELL")
	if err != nil {
		return nil, fmt.Errorf("executor: get SELL price for %s: %w", assetID, err)
	}
	price, err := strconv.ParseFloat(priceResp.Price, 64)
	if err != nil || price <= 0 {
		return nil, fmt.Errorf("executor: invalid SELL price %q for %s", priceResp.Price, assetID)
	}

	req, err := e.buildOrderRequest(assetID, price, sizeShares, auth.Sell)
	if err != nil {
		return nil, fmt.Errorf("executor: build SELL order: %w", err)
	}

	resp, err := e.clob.CreateOrder(req)
	if err != nil {
		return nil, fmt.Errorf("executor: CreateOrder SELL: %w", err)
	}
	if !resp.Success {
		e.logOrder(resp.OrderID, false)
		return nil, fmt.Errorf("executor: SELL order rejected: %s", resp.ErrorMsg)
	}

	pnl := (price - avgBuyPrice) * sizeShares

	e.logOrder(resp.OrderID, true)
	e.logger.Info().
		Str("asset_id", assetID).
		Str("order_id", resp.OrderID).
		Float64("sell_price", price).
		Float64("pnl", pnl).
		Msg("closed copy position")

	return &CloseResult{OrderID: resp.OrderID, Price: price, PnL: pnl}, nil
}

// PlaceLimit places a limit order at a specified price.
// side is "YES" (buy) or "NO" (sell); orderType is "GTC" or "FOK".
// Returns the order ID on success.
func (e *OrderExecutor) PlaceLimit(tokenID, side, orderType string, price, sizeUSD float64) (string, error) {
	if price <= 0 {
		return "", fmt.Errorf("executor: invalid limit price %.4f", price)
	}
	sizeShares := sizeUSD / price

	authSide := auth.Buy
	if side == "NO" {
		authSide = auth.Sell
	}

	req, err := e.buildOrderRequest(tokenID, price, sizeShares, authSide)
	if err != nil {
		return "", fmt.Errorf("executor: build limit order: %w", err)
	}

	ot := clob.OrderTypeGTC
	if orderType == "FOK" {
		ot = clob.OrderTypeFOK
	}
	req.OrderType = ot

	resp, err := e.clob.CreateOrder(req)
	if err != nil {
		return "", fmt.Errorf("executor: CreateOrder limit: %w", err)
	}
	if !resp.Success {
		e.logOrder(resp.OrderID, false)
		return "", fmt.Errorf("executor: limit order rejected: %s", resp.ErrorMsg)
	}

	e.logOrder(resp.OrderID, true)
	e.logger.Info().
		Str("token_id", tokenID).
		Str("side", side).
		Str("order_id", resp.OrderID).
		Float64("price", price).
		Float64("size_usd", sizeUSD).
		Msg("placed limit order")

	return resp.OrderID, nil
}

// buildOrderRequest строит подписанный CreateOrderRequest для CLOB API.
func (e *OrderExecutor) buildOrderRequest(
	assetID string,
	price float64,
	sizeShares float64,
	side auth.OrderSide,
) (*clob.CreateOrderRequest, error) {
	salt, err := auth.RandomSalt()
	if err != nil {
		return nil, fmt.Errorf("executor: generate salt: %w", err)
	}

	makerAddr := common.HexToAddress(e.makerAddr)
	zeroAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// TokenID: token_id из Data API — это uint256 в виде строки
	tokenID := new(big.Int)
	cleanID := strings.TrimPrefix(assetID, "0x")
	// Пробуем сначала как hex, потом как десятичное число
	if _, ok := tokenID.SetString(cleanID, 16); !ok {
		if _, ok := tokenID.SetString(assetID, 10); !ok {
			return nil, fmt.Errorf("executor: cannot parse token_id %q", assetID)
		}
	}

	// Суммы в base units (6 decimals), избегаем потери точности через int64
	priceBig := new(big.Float).SetFloat64(price)
	sizeBig := new(big.Float).SetFloat64(sizeShares)
	decimalsBig := new(big.Float).SetInt64(usdcDecimals)

	toBaseUnits := func(f *big.Float) *big.Int {
		n, _ := new(big.Float).Mul(f, decimalsBig).Int(nil)
		return n
	}

	var makerAmount, takerAmount *big.Int
	if side == auth.Buy {
		// BUY: отдаём USDC (price * size), получаем токены (size)
		makerAmount = toBaseUnits(new(big.Float).Mul(priceBig, sizeBig))
		takerAmount = toBaseUnits(sizeBig)
	} else {
		// SELL: отдаём токены (size), получаем USDC (price * size)
		makerAmount = toBaseUnits(sizeBig)
		takerAmount = toBaseUnits(new(big.Float).Mul(priceBig, sizeBig))
	}

	rawOrder := &auth.RawOrder{
		Salt:          salt,
		Maker:         makerAddr,
		Signer:        makerAddr,
		Taker:         zeroAddr,
		TokenID:       tokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Expiration:    big.NewInt(0),
		Nonce:         big.NewInt(0),
		FeeRateBps:    big.NewInt(defaultFeeRateBps),
		Side:          side,
		SignatureType: auth.EOA,
	}

	sig, err := e.orderSigner.Sign(rawOrder)
	if err != nil {
		return nil, fmt.Errorf("executor: sign order: %w", err)
	}

	sideInt := int(side)

	return &clob.CreateOrderRequest{
		Order: clob.SignedOrder{
			Salt:          salt.String(),
			Maker:         e.makerAddr,
			Signer:        e.makerAddr,
			Taker:         "0x0000000000000000000000000000000000000000",
			TokenID:       assetID,
			MakerAmount:   makerAmount.String(),
			TakerAmount:   takerAmount.String(),
			Expiration:    "0",
			Nonce:         "0",
			FeeRateBps:    strconv.Itoa(defaultFeeRateBps),
			Side:          sideInt,
			SignatureType: int(auth.EOA),
			Signature:     sig,
		},
		Owner:         e.apiKey,
		OrderType:     clob.OrderTypeGTC,
		BuilderApiKey: e.builderAPIKey,
	}, nil
}
