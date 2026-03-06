package copytrading

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/notify"
	"github.com/atlasdev/polytrade-bot/internal/storage"
	"github.com/atlasdev/polytrade-bot/internal/tui"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

// CopyTrader — главный оркестратор подсистемы копитрейдинга.
// Запускает по одному TraderTracker на каждого активного трейдера
// и перезапускает их при изменении конфига.
type CopyTrader struct {
	cfgPath    string
	getCfg     func() *config.CopytradingConfig
	dataClient *data.Client
	executor   *OrderExecutor
	store      storage.CopyTradeStore
	notifier   notify.Notifier
	clobClient *clob.Client
	logger     zerolog.Logger

	bus *tui.EventBus

	mu       sync.Mutex
	trackers map[string]context.CancelFunc // address → cancel
}

// NewCopyTrader создаёт CopyTrader.
//
//   - cfgPath — путь к config.toml (для fsnotify watcher)
//   - getCfg  — функция, возвращающая актуальный CopytradingConfig (вызывается при перезагрузке)
//   - clobClient — для получения нашего баланса USDC
func NewCopyTrader(
	cfgPath string,
	getCfg func() *config.CopytradingConfig,
	dataClient *data.Client,
	executor *OrderExecutor,
	store storage.CopyTradeStore,
	notifier notify.Notifier,
	clobClient *clob.Client,
	log zerolog.Logger,
) *CopyTrader {
	return &CopyTrader{
		cfgPath:    cfgPath,
		getCfg:     getCfg,
		dataClient: dataClient,
		executor:   executor,
		store:      store,
		notifier:   notifier,
		clobClient: clobClient,
		logger:     log.With().Str("component", "copy-trader").Logger(),
		trackers:   make(map[string]context.CancelFunc),
	}
}

// SetBus wires the EventBus so the copier can emit CopytradingTradeMsg events.
func (ct *CopyTrader) SetBus(bus *tui.EventBus) {
	ct.bus = bus
}

// Run запускает копитрейдинг и блокирует до отмены ctx.
// Запускает fsnotify-watcher для горячей перезагрузки конфига.
func (ct *CopyTrader) Run(ctx context.Context) error {
	ct.logger.Info().Msg("copy trader starting")

	// Запустить трекеры для текущего конфига
	cfg := ct.getCfg()
	ct.applyConfig(ctx, cfg)

	// Настроить fsnotify-watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		ct.logger.Warn().Err(err).Msg("failed to create fsnotify watcher, hot-reload disabled")
		<-ctx.Done()
		ct.stopAll()
		return nil
	}
	defer watcher.Close()

	if err := watcher.Add(ct.cfgPath); err != nil {
		ct.logger.Warn().Err(err).Str("path", ct.cfgPath).Msg("failed to watch config file, hot-reload disabled")
	} else {
		ct.logger.Info().Str("path", ct.cfgPath).Msg("watching config for changes")
	}

	for {
		select {
		case <-ctx.Done():
			ct.stopAll()
			ct.logger.Info().Msg("copy trader stopped")
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				ct.logger.Info().Str("file", event.Name).Msg("config changed, reloading traders")
				newCfg := ct.getCfg()
				ct.applyConfig(ctx, newCfg)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			ct.logger.Warn().Err(err).Msg("fsnotify error")
		}
	}
}

// applyConfig сравнивает новый конфиг с текущими трекерами и:
//   - останавливает трекеры для отключённых/удалённых трейдеров
//   - запускает трекеры для новых/включённых трейдеров
func (ct *CopyTrader) applyConfig(ctx context.Context, cfg *config.CopytradingConfig) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	interval := time.Duration(cfg.PollIntervalMs) * time.Millisecond

	// Составим map активных трейдеров из нового конфига
	active := make(map[string]config.TraderConfig)
	for _, t := range cfg.Traders {
		if t.Enabled && t.Address != "" {
			active[t.Address] = t
		}
	}

	// Остановить трекеры, которые больше не нужны
	for addr, cancel := range ct.trackers {
		if _, ok := active[addr]; !ok {
			ct.logger.Info().Str("address", addr).Msg("stopping tracker (removed or disabled)")
			cancel()
			delete(ct.trackers, addr)
		}
	}

	// Запустить новые трекеры
	for addr, traderCfg := range active {
		if _, running := ct.trackers[addr]; running {
			continue
		}
		ct.logger.Info().Str("address", addr).Str("label", traderCfg.Label).Msg("starting tracker")
		ct.startTracker(ctx, traderCfg, interval)
	}
}

// startTracker создаёт и запускает TraderTracker в отдельной горутине.
// Должен вызываться под ct.mu.
func (ct *CopyTrader) startTracker(ctx context.Context, trader config.TraderConfig, interval time.Duration) {
	trackerCtx, cancel := context.WithCancel(ctx)
	ct.trackers[trader.Address] = cancel

	tracker := NewTraderTracker(
		trader,
		ct.dataClient,
		ct.executor,
		ct.store,
		ct.notifier,
		ct.getMyBalance,
		ct.bus,
		ct.logger,
	)

	go func() {
		if err := tracker.Run(trackerCtx, interval); err != nil && trackerCtx.Err() == nil {
			ct.logger.Error().Err(err).Str("address", trader.Address).Msg("tracker exited with error")
		}
	}()
}

// stopAll останавливает все работающие трекеры.
func (ct *CopyTrader) stopAll() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	for addr, cancel := range ct.trackers {
		ct.logger.Info().Str("address", addr).Msg("stopping tracker")
		cancel()
		delete(ct.trackers, addr)
	}
}

// getMyBalance возвращает баланс USDC нашего кошелька через CLOB /balance-allowance.
func (ct *CopyTrader) getMyBalance() (float64, error) {
	ba, err := ct.clobClient.GetBalanceAllowance("COLLATERAL", "")
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	// Баланс приходит в base units (6 decimals)
	raw, err := strconv.ParseFloat(ba.Balance, 64)
	if err != nil {
		return 0, fmt.Errorf("parse balance %q: %w", ba.Balance, err)
	}
	return raw / 1_000_000.0, nil
}
