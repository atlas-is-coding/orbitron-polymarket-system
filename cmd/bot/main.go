package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"

	"github.com/atlasdev/orbitron/internal/api"
	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/api/data"
	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/api/ws"
	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/copytrading"
	"github.com/atlasdev/orbitron/internal/health"
	"github.com/atlasdev/orbitron/internal/i18n"
	"github.com/atlasdev/orbitron/internal/logger"
	"github.com/atlasdev/orbitron/internal/markets"
	"github.com/atlasdev/orbitron/internal/monitor"
	"github.com/atlasdev/orbitron/internal/notify"
	telegramNotify "github.com/atlasdev/orbitron/internal/notify/telegram"
	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/atlasdev/orbitron/internal/storage/sqlite"
	"github.com/atlasdev/orbitron/internal/telegrambot"
	"github.com/atlasdev/orbitron/internal/trading"
	"github.com/atlasdev/orbitron/internal/trading/risk"
	"github.com/atlasdev/orbitron/internal/trading/strategies"
	"github.com/atlasdev/orbitron/internal/license"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/atlasdev/orbitron/internal/updater"
	"github.com/atlasdev/orbitron/internal/wallet"
	"github.com/atlasdev/orbitron/internal/webui"
)

// healthPublisher adapts health.Publisher to tui.EventBus.
type healthPublisher struct{ bus *tui.EventBus }

func (p *healthPublisher) Send(snap health.HealthSnapshot) {
	if p.bus != nil {
		p.bus.Send(tui.HealthSnapshotMsg{Snapshot: snap})
	}
}

type executorAdapter struct{ e *copytrading.OrderExecutor }

func (a *executorAdapter) Open(assetID string, sizeUSD float64, negRisk bool) (*copytrading.OpenResult, error) {
	return a.e.Open(assetID, sizeUSD, negRisk)
}
func (a *executorAdapter) Close(assetID string, sizeShares, avgBuyPrice float64, negRisk bool) (*copytrading.CloseResult, error) {
	return a.e.Close(assetID, sizeShares, avgBuyPrice, negRisk)
}
func (a *executorAdapter) PlaceLimit(tokenID, side, orderType string, price, sizeUSD float64) (string, error) {
	return a.e.PlaceLimit(tokenID, side, orderType, price, sizeUSD)
}

// engineAdapter implements tui.TradingProvider and webui.TradingProvider.
type engineAdapter struct {
	engine  *trading.Engine
	wm      *wallet.Manager
	bus     *tui.EventBus
	ctx     context.Context
	cfgPath string
}

func (a *engineAdapter) WalletIDs() []string { return a.wm.WalletIDs() }
func (a *engineAdapter) WalletLabel(id string) string { return a.wm.WalletLabel(id) }
func (a *engineAdapter) WalletAddress(id string) string { return a.wm.WalletAddress(id) }
func (a *engineAdapter) WalletEnabled(id string) bool { return a.wm.WalletEnabled(id) }
func (a *engineAdapter) WalletStats(id string) (float64, float64, int, int) {
	return a.wm.WalletStats(id)
}
func (a *engineAdapter) AvailableWallets() []string { return a.wm.AvailableWallets() }

func (a *engineAdapter) Remove(id string) error           { return a.wm.Remove(id) }
func (a *engineAdapter) Toggle(id string, enabled bool) error { return a.wm.Toggle(id, enabled) }
func (a *engineAdapter) UpdateLabel(id, label string) error { return a.wm.UpdateLabel(id, label) }
func (a *engineAdapter) SetPrimary(id string) error      { return a.wm.SetPrimary(id) }
func (a *engineAdapter) PlaceOrder(walletID, tokenID, side, orderType string, price, sizeUSD float64, negRisk bool) (string, error) {
	return a.wm.PlaceOrder(walletID, tokenID, side, orderType, price, sizeUSD, negRisk)
}

func (a *engineAdapter) StartStrategy(name string) error {
	err := a.engine.StartStrategy(a.ctx, name)
	if err == nil && a.bus != nil {
		a.bus.Send(tui.StrategiesUpdateMsg{Rows: trading.GetStrategyRows(a.engine, a.wm)})
	}
	return err
}

func (a *engineAdapter) StopStrategy(name string) error {
	err := a.engine.StopStrategy(name)
	if err == nil && a.bus != nil {
		a.bus.Send(tui.StrategiesUpdateMsg{Rows: trading.GetStrategyRows(a.engine, a.wm)})
	}
	return err
}

func (a *engineAdapter) SetStrategyWallets(name string, walletIDs []string) error {
	execAdapters := make(map[string]interface{})
	for _, wid := range walletIDs {
		inst, ok := a.wm.Get(wid)
		if !ok {
			return fmt.Errorf("wallet %s not found", wid)
		}
		if inst.Executor == nil {
			return fmt.Errorf("wallet %s has no executor", wid)
		}
		execAdapters[wid] = &executorAdapter{inst.Executor}
	}

	// Update in-memory config for persistence
	cfg, err := config.Load(a.cfgPath)
	if err == nil {
		updated := false
		switch name {
		case "arbitrage":
			cfg.Trading.Strategies.Arbitrage.WalletIDs = walletIDs
			updated = true
		case "market_making":
			cfg.Trading.Strategies.MarketMaking.WalletIDs = walletIDs
			updated = true
		case "positive_ev":
			cfg.Trading.Strategies.PositiveEV.WalletIDs = walletIDs
			updated = true
		case "riskless_rate":
			cfg.Trading.Strategies.RisklessRate.WalletIDs = walletIDs
			updated = true
		case "fade_chaos":
			cfg.Trading.Strategies.FadeChaos.WalletIDs = walletIDs
			updated = true
		case "cross_market":
			cfg.Trading.Strategies.CrossMarket.WalletIDs = walletIDs
			updated = true
		}
		if updated {
			config.Save(a.cfgPath, cfg)
		}
	}

	err = a.engine.SetStrategyWallets(name, walletIDs, execAdapters)
	if err == nil && a.bus != nil {
		a.bus.Send(tui.StrategiesUpdateMsg{Rows: trading.GetStrategyRows(a.engine, a.wm)})
	}
	return err
}

func (a *engineAdapter) Strategies() []trading.Strategy {
	return a.engine.Strategies()
}

func (a *engineAdapter) CancelOrder(id string) error {
	for _, inst := range a.wm.Wallets() {
		if inst.TradesMon != nil && inst.Cfg.Enabled {
			return inst.TradesMon.CancelOrder(id)
		}
	}
	return fmt.Errorf("no active trades monitor")
}

func (a *engineAdapter) CancelAllOrders() error {
	for _, inst := range a.wm.Wallets() {
		if inst.TradesMon != nil && inst.Cfg.Enabled {
			return inst.TradesMon.CancelAllOrders()
		}
	}
	return fmt.Errorf("no active trades monitor")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// --- Flags ---
	cfgPath := flag.String("config", "config.toml", "path to config file")
	noTUI := flag.Bool("no-tui", false, "disable TUI, use plain log output (headless/CI)")
	flag.Parse()

	// --- Initial setup (wizard) if config.toml does not exist ---
	if _, err := os.Stat(*cfgPath); os.IsNotExist(err) && !*noTUI {
		p := tea.NewProgram(tui.NewWizardModel(80, 24, *cfgPath), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("wizard: %w", err)
		}
		if _, err := os.Stat(*cfgPath); os.IsNotExist(err) {
			return fmt.Errorf("wizard completed without creating config")
		}
	}

	// --- Configuration ---
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// --- Proxy dialer (nil when proxy disabled) ---
	proxyDial, err := api.BuildDialer(cfg.Proxy)
	if err != nil {
		return fmt.Errorf("proxy: %w", err)
	}

	// --- Interface language ---
	i18n.SetLanguage(cfg.UI.Language)

	// --- File logger (if log.file is set) ---
	var logFileCloser func()
	var fileWriter io.Writer
	if cfg.Log.File != "" {
		lf, ferr := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "warn: cannot open log file %q: %v\n", cfg.Log.File, ferr)
		} else {
			fileWriter = lf
			logFileCloser = func() { lf.Close() }
		}
	}
	defer func() {
		if logFileCloser != nil {
			logFileCloser()
		}
	}()

	// --- EventBus + LogWriter (TUI mode or WebUI) ---
	var bus *tui.EventBus
	var log zerolog.Logger
	if !*noTUI || cfg.WebUI.Enabled {
		bus = tui.NewEventBus()
		lw := tui.NewLogWriter(bus)
		var w io.Writer = lw
		if fileWriter != nil {
			w = io.MultiWriter(lw, fileWriter)
		}
		log = logger.NewWithWriter(cfg.Log.Level, cfg.Log.Format, w)
	} else {
		if fileWriter != nil {
			log = logger.NewWithWriter(cfg.Log.Level, cfg.Log.Format, io.MultiWriter(os.Stdout, fileWriter))
		} else {
			log = logger.New(cfg.Log.Level, cfg.Log.Format)
		}
	}
	if proxyDial != nil {
		log.Info().Str("type", cfg.Proxy.Type).Str("addr", cfg.Proxy.Addr).Msg("proxy enabled")
	}

	// --- Nexus State Manager ---
	nx := tui.NewNexus()
	if bus != nil {
		tap := bus.Tap()
		go func() {
			for msg := range tap {
				nx.Handle(msg)
			}
		}()
	}

	// Load Builder Program credentials (non-fatal: bot runs without them).
	builderCreds, licenseErr := license.Load()
	if licenseErr != nil {
		log.Warn().Err(licenseErr).Msg("builder credentials unavailable — Builder features disabled")
	} else if builderCreds != nil {
		log.Info().Str("key", builderCreds.APIKey[:4]+"***").Msg("builder credentials loaded")
	}
	// builderCreds используются ниже при инициализации кошельков.

	// --- Geoblock check ---
	if geo, geoErr := health.CheckGeoblock(proxyDial); geoErr != nil {
		log.Warn().Err(geoErr).Msg("geoblock check failed (continuing)")
	} else if geo.Blocked {
		log.Warn().
			Str("country", geo.Country).
			Str("region", geo.Region).
			Str("ip", geo.IP).
			Msg("⚠ trading blocked in your region — configure [proxy] in config.toml to bypass")
	} else {
		log.Info().Str("country", geo.Country).Str("ip", geo.IP).Msg("geoblock check passed")
	}

	log.Info().Str("config", *cfgPath).Msg(i18n.T().LogBotStarting)

	// --- Wallet Manager ---
	wm := wallet.NewManager(bus)
	wm.SetDialer(proxyDial)
	wm.SetLogger(log)

	// --- HTTP clients ---
	clobHTTP := api.NewClientWithDialer(cfg.API.ClobURL, cfg.API.TimeoutSec, cfg.API.MaxRetries, proxyDial)
	gammaHTTP := api.NewClientWithDialer(cfg.API.GammaURL, cfg.API.TimeoutSec, cfg.API.MaxRetries, proxyDial)
	dataHTTP := api.NewClientWithDialer(cfg.API.DataURL, cfg.API.TimeoutSec, cfg.API.MaxRetries, proxyDial)

	// --- API clients (shared/public) ---
	gammaClient := gamma.NewClient(gammaHTTP)
	dataClient := data.NewClient(dataHTTP)

	// --- WebSocket ---
	wsClient := ws.NewClient(cfg.API.WSURL, log)
	if proxyDial != nil {
		wsClient.WithDialer(func(network, addr string) (net.Conn, error) {
			return proxyDial(addr)
		})
	}

	// --- Notifier ---
	var notifier notify.Notifier = &notify.NoopNotifier{}
	if cfg.Telegram.Enabled {
		notifier = telegramNotify.New(cfg.Telegram.BotToken, cfg.Telegram.AdminChatID)
		log.Info().Msg(i18n.T().LogTelegramEnabled)
	}

	// --- Storage (SQLite) ---
	var db *sqlite.DB
	if cfg.Database.Enabled {
		db, err = sqlite.Open(cfg.Database.Path)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()
		log.Info().Str("path", cfg.Database.Path).Msg(i18n.T().LogDatabaseOpened)
	}

	// --- Build wallet instances ---
	pubClobClient := clob.NewClient(clobHTTP, nil)
	for _, wCfg := range cfg.Wallets {
		if !wCfg.Enabled {
			wm.AddInactive(wCfg)
			continue
		}
		var addr string
		var l1 *auth.L1Signer
		if wCfg.PrivateKey != "" {
			l1, err = auth.NewL1Signer(wCfg.PrivateKey)
			if err != nil {
				log.Warn().Err(err).Str("wallet", wCfg.Label).Msg("l1 signer failed, skipping wallet")
				wm.AddInactive(wCfg)
				continue
			}
			addr = l1.Address()
		}
		l2 := &auth.L2Credentials{
			APIKey:     wCfg.APIKey,
			APISecret:  wCfg.APISecret,
			Passphrase: wCfg.Passphrase,
		}
		if l1 != nil {
			l2.Address = l1.Address()
		}
		if l2.APIKey == "" && l1 != nil {
			derived, deriveErr := pubClobClient.DeriveAPIKey(l1)
			if deriveErr != nil {
				log.Warn().Err(deriveErr).Str("wallet", wCfg.Label).Msg("auto-derive api_key failed")
			} else {
				l2.APIKey = derived.APIKey
				l2.APISecret = derived.APISecret
				l2.Passphrase = derived.Passphrase
				log.Info().Str("wallet", wCfg.Label).Str("address", addr).Msg("api_key auto-derived from private_key")
			}
		}
		if l2.APIKey == "" {
			log.Warn().Str("wallet", wCfg.Label).Msg("wallet has no api_key, skipping")
			wm.AddInactive(wCfg)
			continue
		}
		if addr != "" {
			log.Info().Str("wallet", wCfg.Label).Str("address", addr).Msg("wallet initialized")
		}
		wClobClient := clob.NewClient(clobHTTP, l2)

		// Subscribe WebSocket user events for this wallet
		wsClient.Subscribe(ws.UserSubscription(l2), func(msg *ws.Message) {
			log.Debug().Str("event", msg.EventType).Msg(i18n.T().LogWSUserEvent)
		})

		inst := &wallet.WalletInstance{
			Cfg:        wCfg,
			Address:    addr,
			L2:         l2,
			ClobClient: wClobClient,
			Stats:      &wallet.WalletStats{},
		}
		if l1 != nil {
			orderSigner := auth.NewOrderSigner(l1, wCfg.ChainID, wCfg.NegRisk)
			inst.Executor = copytrading.NewOrderExecutor(wClobClient, orderSigner, l2.APIKey, addr, log)
			if builderCreds != nil {
				inst.Executor.WithBuilderKey(builderCreds.APIKey)
			}
		}
		if cfg.Monitor.Trades.Enabled {
			tm := monitor.NewTradesMonitor(wClobClient, dataClient, notifier, &cfg.Monitor.Trades, log, addr, db)
			if bus != nil {
				tm.SetBus(bus)
			}
			inst.TradesMon = tm
		}
		if cfg.Copytrading.Enabled && db != nil && l1 != nil {
			ct := copytrading.NewCopyTrader(
				*cfgPath,
				func() *config.CopytradingConfig { return &cfg.Copytrading },
				dataClient,
				inst.Executor,
				db,
				notifier,
				wClobClient,
				log,
			)
			if bus != nil {
				ct.SetBus(bus)
			}
			inst.CopyTrader = ct
		}
		wm.AddActive(inst)
	}

	// --- Context with graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Trading Engine ---
	engine := trading.NewEngine(log, wm)

	// --- Risk Manager ---
	riskMgr := risk.NewManager(cfg.Trading.Risk)

	// --- Trading Strategies ---
	// Helper to find executor by wallet ID or primary (returns the first available).
	getExecutor := func(wids []string) strategies.Executor {
		if len(wids) == 0 {
			// Fallback to primary
			for _, id := range wm.WalletIDs() {
				inst, ok := wm.Get(id)
				if ok && inst.Cfg.Primary && inst.Cfg.Enabled && inst.Executor != nil {
					return &executorAdapter{inst.Executor}
				}
			}
			return nil
		}
		for _, wid := range wids {
			if wid == "" {
				continue
			}
			inst, ok := wm.Get(wid)
			if ok && inst.Cfg.Enabled && inst.Executor != nil {
				return &executorAdapter{inst.Executor}
			}
		}
		return nil
	}

	sc := cfg.Trading.Strategies
	engine.Register(strategies.NewArbitrageStrategy(
		gammaClient, getExecutor(sc.Arbitrage.WalletIDs), notifier, bus, riskMgr, sc.Arbitrage, log,
	))
	engine.Register(strategies.NewMarketMakingStrategy(
		gammaClient, pubClobClient, getExecutor(sc.MarketMaking.WalletIDs), notifier, bus, riskMgr, sc.MarketMaking, log,
	))
	engine.Register(strategies.NewPositiveEVStrategy(
		gammaClient, getExecutor(sc.PositiveEV.WalletIDs), notifier, bus, riskMgr, sc.PositiveEV, log,
	))
	engine.Register(strategies.NewRisklessRateStrategy(
		gammaClient, getExecutor(sc.RisklessRate.WalletIDs), notifier, bus, riskMgr, sc.RisklessRate, log,
	))
	engine.Register(strategies.NewFadeTheChaosStrategy(
		gammaClient, getExecutor(sc.FadeChaos.WalletIDs), notifier, bus, riskMgr, sc.FadeChaos, log,
	))
	engine.Register(strategies.NewCrossMarketStrategy(
		gammaClient, getExecutor(sc.CrossMarket.WalletIDs), notifier, bus, riskMgr, sc.CrossMarket, log,
	))

	// --- Market Monitor ---
	mon := monitor.New(gammaClient, notifier, &cfg.Monitor, log)
	if db != nil {
		mon.WithStore(db)
	}

	// --- Markets Cache ---
	var mktCache storage.MarketCacheStore
	if db != nil {
		mktCache = db
	}

	// --- Markets Service ---
	var marketsService *markets.Service
	if bus != nil || cfg.WebUI.Enabled {
		marketsService = markets.NewService(gammaClient, bus, mktCache).WithLogger(&log)
	} else {
		marketsService = markets.NewService(gammaClient, nil, mktCache).WithLogger(&log)
	}

	// --- Health Service ---
	healthSvc := health.New(
		health.Endpoints{
			ClobURL:  cfg.API.ClobURL,
			GammaURL: cfg.API.GammaURL,
			DataURL:  cfg.API.DataURL,
			WSURL:    cfg.API.WSURL,
		},
		proxyDial,
		&healthPublisher{bus: bus},
		log,
	)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg(i18n.T().LogShutdownSignal)
		cancel()
	}()

	// --- Start subsystems ---
	errCh := make(chan error, 8)

	startSubsystem := func(name string, fn func() error) {
		go func() {
			if err := fn(); err != nil && ctx.Err() == nil {
				errCh <- fmt.Errorf("%s: %w", name, err)
			}
		}()
		if bus != nil {
			bus.Send(tui.SubsystemStatusMsg{Name: name, Active: true})
		}
	}

	startSubsystem("WebSocket", func() error { return wsClient.Run(ctx) })
	startSubsystem("Markets", func() error { return marketsService.Run(ctx) })
	startSubsystem("Health", func() error {
		healthSvc.Start(ctx)
		return nil
	})

	if cfg.Monitor.Enabled {
		startSubsystem("Monitor", func() error { return mon.Run(ctx) })
	}

	if cfg.Trading.Enabled {
		startSubsystem("Trading Engine", func() error { return engine.Start(ctx) })
	}

	// --- Start per-wallet subsystems ---
	for _, inst := range wm.Wallets() {
		if !inst.Cfg.Enabled {
			continue
		}
		label := inst.Cfg.Label
		if inst.TradesMon != nil {
			tm := inst.TradesMon
			startSubsystem("Trades Monitor ["+label+"]", func() error { return tm.Run(ctx) })
		}
		if inst.CopyTrader != nil {
			ct := inst.CopyTrader
			startSubsystem("Copytrading ["+label+"]", func() error { return ct.Run(ctx) })
		}
	}

	// --- Stats poller (wallet balance / P&L via Data API) ---
	go wm.RunStatsPoller(ctx, dataClient, 30*time.Second, db)

	adapter := &engineAdapter{engine: engine, wm: wm, bus: bus, ctx: ctx, cfgPath: *cfgPath}

	// --- Telegram Bot (interactive) ---
	// Initialised before subsystems so it can be started alongside them.
	// tgBot may be nil if bot_token is empty or init fails.
	var tgBot *telegrambot.Bot
	if cfg.Telegram.Enabled && cfg.Telegram.BotToken != "" {
		var cancelerForBot telegrambot.OrderCanceler
		for _, inst := range wm.Wallets() {
			if inst.TradesMon != nil && inst.Cfg.Enabled {
				cancelerForBot = inst.TradesMon
				break
			}
		}
		tgBot, err = telegrambot.New(cfg, *cfgPath, bus, cancelerForBot, wm, wm, marketsService, wm, &log)
		if err != nil {
			log.Warn().Err(err).Msg("telegram bot init failed, continuing without it")
			tgBot = nil
		}
	}

	if tgBot != nil {
		startSubsystem("Telegram Bot", func() error { return tgBot.Run(ctx) })
	}

	// --- Update checker ---
	// Uses updater.Dir() as the single source of truth for the working directory.
	// notifier is the Telegram notify.Notifier already configured above.
	pending := updater.NewPending(updater.Dir())
	updateNotifier := updater.NewNotifier(bus, notifier)
	go updater.Start(ctx, engine.IsIdle, updateNotifier, pending)

	if cfg.WebUI.Enabled && bus != nil {
		var cancelerForWeb webui.OrderCanceler
		for _, inst := range wm.Wallets() {
			if inst.TradesMon != nil && inst.Cfg.Enabled {
				cancelerForWeb = inst.TradesMon
				break
			}
		}
		webServer := webui.New(cfg, *cfgPath, bus, nx, cancelerForWeb, wm, marketsService, adapter, adapter, &log)
		startSubsystem("Web UI", func() error { return webServer.Run(ctx) })
	}

	// --- TUI mode ---
	if !*noTUI && bus != nil {
		// ConfigWatcher — hot reload via fsnotify
		watcher, _ := config.NewWatcher(*cfgPath, func(newCfg *config.Config) {
			bus.Send(tui.ConfigReloadedMsg{Config: newCfg})
			bus.Send(tui.StrategiesUpdateMsg{Rows: trading.GetStrategyRows(engine, wm)})
		})
		go watcher.Run(ctx)

		// Emit initial strategies
		bus.Send(tui.StrategiesUpdateMsg{Rows: trading.GetStrategyRows(engine, wm)})

		// Start TUI
		rootModel := tui.NewRootModel(cfg, *cfgPath, bus, nx, 0, 0, nil, adapter)


		// Show first active wallet address
		for _, inst := range wm.Wallets() {
			if inst.Address != "" && inst.Cfg.Enabled {
				rootModel.SetWallet(inst.Address)
				break
			}
		}

		p := tea.NewProgram(rootModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("tui: %w", err)
		}
		// Goodbye message (printed after alt screen is restored)
		fmt.Println("\n  ◈ orbitron — shutdown complete. Goodbye!")
		cancel()
		return nil
	}

	// --- Headless mode ---
	log.Info().Msg(i18n.T().LogBotRunning)
	select {
	case <-ctx.Done():
		log.Info().Msg(i18n.T().LogShuttingDown)
	case err := <-errCh:
		log.Error().Err(err).Msg(i18n.T().LogFatalError)
		cancel()
		return err
	}

	wsClient.Close()
	engine.Stop()
	log.Info().Msg(i18n.T().LogBye)
	return nil
}
