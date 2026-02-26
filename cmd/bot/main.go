package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	"github.com/atlasdev/polytrade-bot/internal/trading"
	"github.com/atlasdev/polytrade-bot/internal/tui"
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

	// --- EventBus + LogWriter (TUI режим) ---
	var bus *tui.EventBus
	var log zerolog.Logger
	if !*noTUI {
		bus = tui.NewEventBus()
		lw := tui.NewLogWriter(bus)
		log = logger.NewWithWriter(cfg.Log.Level, cfg.Log.Format, lw)
	} else {
		log = logger.New(cfg.Log.Level, cfg.Log.Format)
	}
	log.Info().Str("config", *cfgPath).Msg(i18n.T().LogBotStarting)

	// --- HTTP клиенты ---
	clobHTTP := api.NewClient(cfg.API.ClobURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)
	gammaHTTP := api.NewClient(cfg.API.GammaURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)
	dataHTTP := api.NewClient(cfg.API.DataURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)

	// --- Auth (L2) ---
	var l2Creds *auth.L2Credentials
	var walletAddr string
	if cfg.Auth.APIKey != "" {
		l2Creds = &auth.L2Credentials{
			APIKey:     cfg.Auth.APIKey,
			APISecret:  cfg.Auth.APISecret,
			Passphrase: cfg.Auth.Passphrase,
		}
		if cfg.Auth.PrivateKey != "" {
			l1, err := auth.NewL1Signer(cfg.Auth.PrivateKey)
			if err != nil {
				return fmt.Errorf("l1 signer: %w", err)
			}
			l2Creds.Address = l1.Address()
			walletAddr = l1.Address()
			log.Info().Str("address", l2Creds.Address).Msg(i18n.T().LogL1Initialized)
		}
	}

	// --- API клиенты ---
	clobClient := clob.NewClient(clobHTTP, l2Creds)
	gammaClient := gamma.NewClient(gammaHTTP)
	dataClient := data.NewClient(dataHTTP)

	// --- WebSocket ---
	wsClient := ws.NewClient(cfg.API.WSURL, log)
	if l2Creds != nil {
		wsClient.Subscribe(ws.UserSubscription(l2Creds), func(msg *ws.Message) {
			log.Debug().Str("event", msg.EventType).Msg(i18n.T().LogWSUserEvent)
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

	// --- Trading Engine ---
	engine := trading.NewEngine(log)
	_ = clobClient

	// --- Market Monitor ---
	mon := monitor.New(gammaClient, notifier, &cfg.Monitor, log)

	// --- Trades Monitor ---
	tradesMon := monitor.NewTradesMonitor(clobClient, dataClient, notifier, &cfg.Monitor.Trades, log)

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

	if cfg.Monitor.Trades.Enabled {
		if l2Creds == nil {
			log.Warn().Msg(i18n.T().LogTradesMonitorSkip)
		} else {
			log.Info().Msg(i18n.T().LogTradesMonitorEnabled)
			startSubsystem("Trades Monitor", func() error { return tradesMon.Run(ctx) })
		}
	}

	if cfg.Trading.Enabled {
		startSubsystem("Trading Engine", func() error { return engine.Start(ctx) })
	}

	if cfg.Copytrading.Enabled {
		if l2Creds == nil || cfg.Auth.PrivateKey == "" {
			log.Warn().Msg(i18n.T().LogCopytradingSkipL2)
		} else if db == nil {
			log.Warn().Msg(i18n.T().LogCopytradingSkipDB)
		} else {
			l1, err := auth.NewL1Signer(cfg.Auth.PrivateKey)
			if err != nil {
				return fmt.Errorf("copytrading l1 signer: %w", err)
			}
			orderSigner := auth.NewOrderSigner(l1, cfg.Auth.ChainID, cfg.Trading.NegRisk)
			executor := copytrading.NewOrderExecutor(clobClient, orderSigner, cfg.Auth.APIKey, l2Creds.Address, log)
			copyTrader := copytrading.NewCopyTrader(
				*cfgPath,
				func() *config.CopytradingConfig { return &cfg.Copytrading },
				dataClient,
				executor,
				db,
				notifier,
				clobClient,
				log,
			)
			log.Info().Int("traders", len(cfg.Copytrading.Traders)).Msg(i18n.T().LogCopytradingEnabled)
			startSubsystem("Copytrading", func() error { return copyTrader.Run(ctx) })
		}
	}

	// --- TUI режим ---
	if !*noTUI && bus != nil {
		// ConfigWatcher — hot reload через fsnotify
		watcher, _ := config.NewWatcher(*cfgPath, func(newCfg *config.Config) {
			bus.Send(tui.ConfigReloadedMsg{Config: newCfg})
		})
		go watcher.Run(ctx)

		// Отправляем адрес кошелька
		if walletAddr != "" {
			bus.Send(tui.SubsystemStatusMsg{Name: "WebSocket", Active: true}) // повторно чтобы не потерялось
		}

		// Запускаем TUI
		appModel := tui.NewAppModel(cfg, *cfgPath, bus, 0, 0, nil)
		appModel.SetWallet(walletAddr)

		p := tea.NewProgram(appModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("tui: %w", err)
		}
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
