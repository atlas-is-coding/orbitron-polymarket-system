package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"

	"github.com/atlasdev/polytrade-bot/internal/api"
	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/api/ws"
	"github.com/atlasdev/polytrade-bot/internal/auth"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/copytrading"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
	"github.com/atlasdev/polytrade-bot/internal/logger"
	"github.com/atlasdev/polytrade-bot/internal/monitor"
	"github.com/atlasdev/polytrade-bot/internal/notify"
	telegramNotify "github.com/atlasdev/polytrade-bot/internal/notify/telegram"
	"github.com/atlasdev/polytrade-bot/internal/storage/sqlite"
	"github.com/atlasdev/polytrade-bot/internal/telegrambot"
	"github.com/atlasdev/polytrade-bot/internal/trading"
	"github.com/atlasdev/polytrade-bot/internal/tui"
	"github.com/atlasdev/polytrade-bot/internal/wallet"
	"github.com/atlasdev/polytrade-bot/internal/webui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// --- Флаги ---
	cfgPath := flag.String("config", "config.toml", "path to config file")
	noTUI := flag.Bool("no-tui", false, "disable TUI, use plain log output (headless/CI)")
	flag.Parse()

	// --- Первичная настройка (wizard) если config.toml не существует ---
	if _, err := os.Stat(*cfgPath); os.IsNotExist(err) && !*noTUI {
		p := tea.NewProgram(tui.NewWizardModel(80, 24, *cfgPath), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("wizard: %w", err)
		}
		if _, err := os.Stat(*cfgPath); os.IsNotExist(err) {
			return fmt.Errorf("wizard completed without creating config")
		}
	}

	// --- Конфиг ---
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// --- Язык интерфейса ---
	i18n.SetLanguage(cfg.UI.Language)

	// --- EventBus + LogWriter (TUI режим или WebUI) ---
	var bus *tui.EventBus
	var log zerolog.Logger
	if !*noTUI || cfg.WebUI.Enabled {
		bus = tui.NewEventBus()
		lw := tui.NewLogWriter(bus)
		log = logger.NewWithWriter(cfg.Log.Level, cfg.Log.Format, lw)
	} else {
		log = logger.New(cfg.Log.Level, cfg.Log.Format)
	}
	log.Info().Str("config", *cfgPath).Msg(i18n.T().LogBotStarting)

	// --- Wallet Manager ---
	wm := wallet.NewManager(bus)

	// --- HTTP клиенты ---
	clobHTTP := api.NewClient(cfg.API.ClobURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)
	gammaHTTP := api.NewClient(cfg.API.GammaURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)
	dataHTTP := api.NewClient(cfg.API.DataURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)

	// --- API клиенты (shared/public) ---
	gammaClient := gamma.NewClient(gammaHTTP)
	dataClient := data.NewClient(dataHTTP)

	// --- WebSocket ---
	wsClient := ws.NewClient(cfg.API.WSURL, log)

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
		if cfg.Monitor.Trades.Enabled {
			tm := monitor.NewTradesMonitor(wClobClient, dataClient, notifier, &cfg.Monitor.Trades, log)
			if bus != nil {
				tm.SetBus(bus)
			}
			inst.TradesMon = tm
		}
		if cfg.Copytrading.Enabled && db != nil && l1 != nil {
			orderSigner := auth.NewOrderSigner(l1, wCfg.ChainID, wCfg.NegRisk)
			executor := copytrading.NewOrderExecutor(wClobClient, orderSigner, l2.APIKey, addr, log)
			ct := copytrading.NewCopyTrader(
				*cfgPath,
				func() *config.CopytradingConfig { return &cfg.Copytrading },
				dataClient,
				executor,
				db,
				notifier,
				wClobClient,
				log,
			)
			inst.CopyTrader = ct
		}
		wm.AddActive(inst)
	}

	// --- Trading Engine ---
	engine := trading.NewEngine(log)

	// --- Market Monitor ---
	mon := monitor.New(gammaClient, notifier, &cfg.Monitor, log)

	// --- Context с graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg(i18n.T().LogShutdownSignal)
		cancel()
	}()

	// --- Запуск подсистем ---
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
	go wm.RunStatsPoller(ctx, dataClient, 30*time.Second)

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
		tgBot, err = telegrambot.New(cfg, *cfgPath, bus, cancelerForBot, wm, &log)
		if err != nil {
			log.Warn().Err(err).Msg("telegram bot init failed, continuing without it")
			tgBot = nil
		}
	}

	if tgBot != nil {
		startSubsystem("Telegram Bot", func() error { return tgBot.Run(ctx) })
	}

	if cfg.WebUI.Enabled && bus != nil {
		var cancelerForWeb webui.OrderCanceler
		for _, inst := range wm.Wallets() {
			if inst.TradesMon != nil && inst.Cfg.Enabled {
				cancelerForWeb = inst.TradesMon
				break
			}
		}
		webServer := webui.New(cfg, *cfgPath, bus, cancelerForWeb, wm, wm, nil, &log)
		startSubsystem("Web UI", func() error { return webServer.Run(ctx) })
	}

	// --- TUI режим ---
	if !*noTUI && bus != nil {
		// ConfigWatcher — hot reload через fsnotify
		watcher, _ := config.NewWatcher(*cfgPath, func(newCfg *config.Config) {
			bus.Send(tui.ConfigReloadedMsg{Config: newCfg})
		})
		go watcher.Run(ctx)

		// Запускаем TUI
		rootModel := tui.NewRootModel(cfg, *cfgPath, bus, 0, 0, nil, wm)

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
		fmt.Println("\n  ◈ polytrade-bot — shutdown complete. Goodbye!")
		cancel()
		return nil
	}

	// --- Headless режим ---
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
