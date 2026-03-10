# Trading Strategies Design

**Date**: 2026-03-10
**Status**: Approved
**Source**: TRADING_STRATEGIES.md (reddit + ctrlpoly.xyz)

---

## Scope

Implement Group A strategies (fully automatable using only Polymarket data). Each strategy supports:
- Signal-only mode (alert without order execution)
- Real order execution via CLOB API (requires L2 auth)
- Telegram + TUI alerts on trigger

---

## Strategies

### 1. ArbitrageStrategy
Monitors orderbook prices. When YES price + NO price < $1.00, buys both sides simultaneously, guaranteeing profit regardless of outcome.

**Config**: `[trading.strategies.arbitrage]`
**Signal**: Instant market buy of both tokens
**Key params**: `min_profit_usd`, `max_position_usd`, `execute_orders`

### 2. MarketMakingStrategy
Posts limit orders (bid + ask) around the current midpoint price, collecting spread. Rebalances periodically when price moves away.

**Config**: `[trading.strategies.market_making]`
**Signal**: Limit orders at midpoint ± spread_pct
**Key params**: `spread_pct`, `max_position_usd`, `rebalance_interval_sec`, `execute_orders`

### 3. PositiveEVStrategy
Scans markets for mispriced outcomes: markets where the price is significantly below the true probability (e.g., price 0.70 for a 90%+ likely outcome). Uses liquidity-weighted signals.

**Config**: `[trading.strategies.positive_ev]`
**Signal**: Buy YES when edge > min_edge_pct
**Key params**: `min_edge_pct`, `min_liquidity_usd`, `max_position_usd`, `execute_orders`

### 4. RisklessRateStrategy
Targets long-duration markets (>30 days) where NO trades above fair value. Absurd events (aliens, etc.) should price NO at ~97-98 cents, not lower. Buys NO when overpriced.

**Config**: `[trading.strategies.riskless_rate]`
**Signal**: Buy NO on long-duration markets above price threshold
**Key params**: `min_duration_days`, `max_no_price`, `max_position_usd`, `execute_orders`

### 5. FadeTheChaosStrategy
Detects emotional price spikes: when YES jumps >10% in a single poll interval without fundamental change (measured by volume/liquidity ratio), signals a contrarian NO position.

**Config**: `[trading.strategies.fade_chaos]`
**Signal**: Buy NO after YES spike above threshold
**Key params**: `spike_threshold_pct`, `cooldown_sec`, `max_position_usd`, `execute_orders`

### 6. CrossMarketStrategy *(AI-generated)*
Builds a consistency graph of logically related markets. When Market A is a strict subset of Market B (e.g., "Candidate A wins" ⊆ "Party X wins"), their prices must satisfy `P(A) ≤ P(B)`. Trades violations as arbitrage opportunities.

**Config**: `[trading.strategies.cross_market]`
**Signal**: Alert + proportional position when divergence > min_divergence_pct
**Key params**: `min_divergence_pct`, `max_position_usd`, `execute_orders`

---

## Configuration Schema

```toml
[trading.strategies.arbitrage]
enabled = true
min_profit_usd = 0.50
max_position_usd = 100.0
execute_orders = true

[trading.strategies.market_making]
enabled = false
spread_pct = 2.0
max_position_usd = 200.0
rebalance_interval_sec = 30
execute_orders = true

[trading.strategies.positive_ev]
enabled = true
min_edge_pct = 5.0
min_liquidity_usd = 5000.0
max_position_usd = 50.0
execute_orders = false

[trading.strategies.riskless_rate]
enabled = true
min_duration_days = 30
max_no_price = 0.05
max_position_usd = 50.0
execute_orders = false

[trading.strategies.fade_chaos]
enabled = true
spike_threshold_pct = 10.0
cooldown_sec = 300
max_position_usd = 50.0
execute_orders = false

[trading.strategies.cross_market]
enabled = true
min_divergence_pct = 5.0
max_position_usd = 75.0
execute_orders = false

[trading.risk]
stop_loss_pct = 20.0
take_profit_pct = 50.0
max_daily_loss_usd = 100.0
```

---

## Architecture

### File Layout

```
internal/trading/
  strategy.go              # existing Strategy interface
  engine.go                # existing Engine
  strategies/
    example.go             # existing example
    arbitrage.go           # new
    market_making.go       # new
    positive_ev.go         # new
    riskless_rate.go       # new
    fade_chaos.go          # new
    cross_market.go        # new (AI-generated)
    arbitrage_test.go      # new
    market_making_test.go  # new
    positive_ev_test.go    # new
    riskless_rate_test.go  # new
    fade_chaos_test.go     # new
    cross_market_test.go   # new
  risk/
    manager.go             # new: RiskManager (circuit breaker, stop-loss, take-profit)
    manager_test.go        # new
```

### Dependencies (injected via constructor)

Each strategy receives:
- `gammaClient` — market scanning
- `clobClient` — orderbook prices
- `executor` — order placement (may be nil → signal-only)
- `notifier` — Telegram alerts
- `bus` — EventBus for TUI
- `cfg` — strategy-specific config section
- `riskMgr` — shared RiskManager
- `log` — zerolog.Logger

### RiskManager

```go
type RiskManager struct {
    cfg        RiskConfig
    dailyLoss  float64
    mu         sync.Mutex
    broken     atomic.Bool   // circuit breaker flag
}

func (r *RiskManager) CanTrade() bool          // false when circuit broken
func (r *RiskManager) RecordPnL(usd float64)   // negative = loss
func (r *RiskManager) ResetDaily()             // called at midnight
func (r *RiskManager) CheckPositions(positions []clob.Position, executor Executor) // stop-loss/take-profit
```

### Alert Flow

1. Strategy detects signal
2. Calls `bus.Send(StrategyAlertMsg{...})` → TUI Tab Logs + WebUI WS event
3. Calls `notifier.Send(ctx, msg)` → Telegram message
4. If `execute_orders && executor != nil && riskMgr.CanTrade()` → places order

### New EventBus Message

```go
type StrategyAlertMsg struct {
    Strategy  string    // "arbitrage", "fade_chaos", etc.
    Market    string    // condition_id
    Question  string    // human-readable market question
    Signal    string    // "BUY_YES", "BUY_NO", "SELL", "MARKET_MAKE"
    Price     float64   // current token price
    EdgePct   float64   // estimated edge
    Reason    string    // human-readable explanation
    Executed  bool      // whether order was placed
    OrderID   string    // if executed
}
```

---

## Risk Management

**Standard approach (B) + simple circuit breaker:**

- `max_position_usd` per strategy — hard cap on any single position
- `stop_loss_pct` — RiskManager polls TradesMonitor positions, closes when loss > threshold
- `take_profit_pct` — closes when profit > threshold
- `max_daily_loss_usd` — circuit breaker: all strategies pause until next day reset

RiskManager runs in its own goroutine, checking positions every `monitor.trades.poll_interval_ms`.

---

## Testing

### Unit Tests (no build tag)
- Each strategy: mock CLOB/Gamma returning crafted data, verify signal detection logic
- RiskManager: circuit breaker activation/reset, stop-loss trigger, take-profit trigger
- No real network calls

### Integration Tests (`-tags=integration`)
- PositiveEVStrategy: real Gamma API scan, verify no panics
- FadeTheChaosStrategy: real Gamma API, verify baseline price recording
- Require `POLY_PRIVATE_KEY` env var

---

## Wiring in main.go

```go
riskMgr := risk.NewManager(cfg.Trading.Risk, log)

if cfg.Trading.Enabled {
    if cfg.Trading.Strategies.Arbitrage.Enabled {
        engine.Register(strategies.NewArbitrage(cfg, clobClient, gammaClient, executor, notifier, bus, riskMgr, log))
    }
    // ... other strategies
    startSubsystem("Trading Engine", func() error { return engine.Start(ctx) })
}
```

---

## Constraints

- Strategies use existing `copytrading.OrderExecutor` interface (or a compatible subset)
- No new external dependencies — only existing Polymarket APIs
- Signal-only mode must work without L2 credentials
- All strategy names match config keys exactly (used for routing in alerts)
