package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/atlasdev/polytrade-bot/internal/api"
	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/api/ws"
	"github.com/atlasdev/polytrade-bot/internal/auth"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/logger"
	"github.com/atlasdev/polytrade-bot/internal/monitor"
	"github.com/atlasdev/polytrade-bot/internal/notify"
	telegramNotify "github.com/atlasdev/polytrade-bot/internal/notify/telegram"
	"github.com/atlasdev/polytrade-bot/internal/trading"
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
	flag.Parse()

	// --- Конфиг ---
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// --- Логгер ---
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info().Str("config", *cfgPath).Msg("polytrade-bot starting")

	// --- HTTP клиенты ---
	clobHTTP := api.NewClient(cfg.API.ClobURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)
	gammaHTTP := api.NewClient(cfg.API.GammaURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)
	dataHTTP := api.NewClient(cfg.API.DataURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)

	// --- Auth (L2) ---
	var l2Creds *auth.L2Credentials
	if cfg.Auth.APIKey != "" {
		l2Creds = &auth.L2Credentials{
			APIKey:     cfg.Auth.APIKey,
			APISecret:  cfg.Auth.APISecret,
			Passphrase: cfg.Auth.Passphrase,
		}

		// Адрес получаем из L1 если есть приватный ключ
		if cfg.Auth.PrivateKey != "" {
			l1, err := auth.NewL1Signer(cfg.Auth.PrivateKey)
			if err != nil {
				return fmt.Errorf("l1 signer: %w", err)
			}
			l2Creds.Address = l1.Address()
			log.Info().Str("address", l2Creds.Address).Msg("L1 signer initialized")
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
			log.Debug().Str("event", msg.EventType).Msg("ws user event")
		})
	}

	// --- Notifier ---
	var notifier notify.Notifier = &notify.NoopNotifier{}
	if cfg.Telegram.Enabled {
		notifier = telegramNotify.New(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
		log.Info().Msg("telegram notifier enabled")
	}

	// --- Trading Engine ---
	engine := trading.NewEngine(log)
	_ = clobClient // engine будет использовать clobClient через стратегии

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
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")
		cancel()
	}()

	// --- Запуск компонентов ---
	errCh := make(chan error, 4)

	// WebSocket клиент
	go func() {
		if err := wsClient.Run(ctx); err != nil && ctx.Err() == nil {
			errCh <- fmt.Errorf("websocket: %w", err)
		}
	}()

	// Мониторинг рынков
	if cfg.Monitor.Enabled {
		go func() {
			if err := mon.Run(ctx); err != nil && ctx.Err() == nil {
				errCh <- fmt.Errorf("monitor: %w", err)
			}
		}()
	}

	// Мониторинг сделок и позиций (требует L2)
	if cfg.Monitor.Trades.Enabled {
		if l2Creds == nil {
			log.Warn().Msg("trades monitor requires L2 credentials (api_key/secret/passphrase), skipping")
		} else {
			log.Info().Msg("trades monitor enabled")
			go func() {
				if err := tradesMon.Run(ctx); err != nil && ctx.Err() == nil {
					errCh <- fmt.Errorf("trades monitor: %w", err)
				}
			}()
		}
	}

	// Торговый движок
	if cfg.Trading.Enabled {
		go func() {
			if err := engine.Start(ctx); err != nil && ctx.Err() == nil {
				errCh <- fmt.Errorf("trading engine: %w", err)
			}
		}()
	}

	log.Info().Msg("bot running. Press Ctrl+C to stop.")

	// Ждём завершения
	select {
	case <-ctx.Done():
		log.Info().Msg("shutting down...")
	case err := <-errCh:
		log.Error().Err(err).Msg("fatal error")
		cancel()
		return err
	}

	wsClient.Close()
	engine.Stop()
	log.Info().Msg("bye!")
	return nil
}
