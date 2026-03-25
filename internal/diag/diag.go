package diag

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/atlasdev/orbitron/internal/api"
	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/api/data"
	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/atlasdev/orbitron/internal/builder"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/license"
	"github.com/atlasdev/orbitron/internal/wallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

// Run выполняет диагностику системы: проверку связи, авторизации и возможность торговли.
func Run(ctx context.Context, cfg *config.Config, log zerolog.Logger) error {
	log.Info().Msg("🚀 Starting diagnostics...")

	// 0. Builder credentials
	log.Info().Msg("--- Checking Builder Credentials ---")
	builderCreds, licErr := license.Load()
	
	builder.NewBuilderKeyValidator(builderCreds, log).Check()
	var builderAPIKey string
	if licErr != nil {
		log.Warn().Err(licErr).Msg("⚠️  Builder credentials unavailable — order will be placed WITHOUT builderApiKey")
	} else if builderCreds != nil {
		builderAPIKey = builderCreds.APIKey
	} else {
		log.Warn().Msg("⚠️  No builder token configured — order will be placed WITHOUT builderApiKey")
	}

	// 1. Клиенты
	clobHTTP := api.NewClient(cfg.API.ClobURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)
	gammaHTTP := api.NewClient(cfg.API.GammaURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)
	dataHTTP := api.NewClient(cfg.API.DataURL, cfg.API.TimeoutSec, cfg.API.MaxRetries)

	clobClient := clob.NewClient(clobHTTP, nil)
	gammaClient := gamma.NewClient(gammaHTTP)
	_ = data.NewClient(dataHTTP) // Check connectivity but don't need the client here

	// 2. Connectivity check
	log.Info().Msg("--- Checking Connectivity ---")
	if err := checkOK(clobHTTP, "CLOB API"); err != nil {
		log.Error().Err(err).Msg("CLOB API connectivity failed")
	} else {
		log.Info().Msg("✅ CLOB API is OK")
	}

	// Data API doesn't have /ok, check /positions with a dummy user
	if resp, err := dataHTTP.Get("/positions?user=0x0000000000000000000000000000000000000000&limit=1", nil); err != nil || resp.StatusCode != 200 {
		if err != nil {
			log.Error().Err(err).Msg("Data API connectivity failed")
		} else {
			log.Error().Int("status", resp.StatusCode).Msg("Data API connectivity failed (non-200)")
		}
	} else {
		log.Info().Msg("✅ Data API is OK")
	}

	// 3. Auth & Balance check
	log.Info().Msg("--- Checking Auth & Balance ---")
	if len(cfg.Wallets) == 0 {
		return fmt.Errorf("no wallets configured")
	}
	w := cfg.Wallets[0]
	l1, err := auth.NewL1Signer(w.PrivateKey)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}
	log.Info().Str("address", l1.Address()).Msg("Wallet loaded")

	creds, err := clobClient.DeriveAPIKey(l1, w.ChainID)
	if err != nil {
		return fmt.Errorf("failed to derive L2 credentials: %w", err)
	}
	log.Info().Str("api_key", creds.APIKey).Msg("✅ L2 credentials derived")

	// Пересоздаём CLOB клиент с credentials
	clobClient = clob.NewClient(clobHTTP, creds)

	ba, err := clobClient.GetBalanceAllowance("COLLATERAL", "")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch balance")
	} else {
		balance, _ := strconv.ParseFloat(ba.Balance, 64)
		log.Info().
			Str("balance", fmt.Sprintf("%.2f USDC", balance/1_000_000)).
			Str("allowance", ba.Allowance).
			Msg("✅ Balance & Allowance fetched")

		if ba.Allowance == "0" || ba.Allowance == "" {
			log.Warn().Msg("⚠️ Allowance is 0 or empty! Attempting to grant infinite allowance...")
			statuses, err := wallet.CheckAllowances(ctx, cfg.API.PolygonRPC, l1.Address())
			if err != nil {
				log.Error().Err(err).Msg("Failed to check on-chain allowances")
			} else {
				err = wallet.GrantMissingAllowances(ctx, cfg.API.PolygonRPC, w.PrivateKey, statuses)
				if err != nil {
					log.Error().Err(err).Msg("Failed to grant allowances (ensure wallet has MATIC for gas)")
				} else {
					log.Info().Msg("✅ Infinite allowance transactions sent! (May take a minute to confirm)")
				}
			}
		}
		}

		// 4. Trade check (Place & Cancel)
		log.Info().Msg("--- Checking Trade (Place & Cancel) ---")
		markets, err := gammaClient.GetMarkets(gamma.MarketsParams{
		Active: ptr(true),
		Closed: ptr(false),
		Limit:  10,
		})
		if err != nil {
		return fmt.Errorf("failed to fetch markets from Gamma: %w", err)
		}

		var targetMarket *gamma.Market
		for _, m := range markets {
		// Use Active as a proxy for tradability in diagnostic
		if m.Active && len(m.ClobTokenIDs) >= 2 {
			targetMarket = &m
			break
		}
		}

		if targetMarket == nil {
		log.Warn().Msg("No active markets found for testing")
		return nil
		}

		tokenID := targetMarket.ClobTokenIDs[0] // YES token
		log.Info().
		Str("market", targetMarket.Question).
		Str("token_id", tokenID).
		Bool("neg_risk", targetMarket.NegRisk).
		Msg("Found test market")

		priceLevel := 0.50

		// Строим ордер на $1.00
		orderID, err := testPlaceOrder(clobClient, l1, creds, tokenID, priceLevel, 1.0, w.ChainID, targetMarket.NegRisk, builderAPIKey, log)
		if err != nil {
		log.Error().Err(err).Msg("❌ Failed to place test order")
		} else {
		log.Info().Str("order_id", orderID).Msg("✅ Test order placed")

		// Сразу отменяем
		cancelResp, err := clobClient.CancelOrder(orderID)
		if err != nil {
			log.Error().Err(err).Msg("❌ Failed to cancel test order")
		} else if cancelResp.Canceled {
			log.Info().Msg("✅ Test order cancelled successfully")
		}
		}
	log.Info().Msg("🏁 Diagnostics complete!")
	return nil
}

func checkOK(client *api.Client, name string) error {
	resp, err := client.Get("/ok", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s returned HTTP %d", name, resp.StatusCode)
	}
	return nil
}

func testPlaceOrder(
	clobClient *clob.Client,
	l1 *auth.L1Signer,
	creds *auth.L2Credentials,
	tokenID string,
	price float64,
	sizeUSD float64,
	chainID int64,
	negRisk bool,
	builderAPIKey string,
	log zerolog.Logger,
) (string, error) {
	signer := auth.NewOrderSigner(l1, chainID, negRisk)
	salt, _ := auth.RandomSalt()

	sizeShares := sizeUSD / price
	
	// Convert to base units (6 decimals)
	priceBig := new(big.Float).SetFloat64(price)
	sizeBig := new(big.Float).SetFloat64(sizeShares)
	decimalsBig := new(big.Float).SetInt64(1_000_000)

	toBaseUnits := func(f *big.Float) *big.Int {
		n, _ := new(big.Float).Mul(f, decimalsBig).Int(nil)
		return n
	}

	makerAmount := toBaseUnits(new(big.Float).Mul(priceBig, sizeBig))
	takerAmount := toBaseUnits(sizeBig)

	tokenIDInt := new(big.Int)
	if _, ok := tokenIDInt.SetString(tokenID, 10); !ok {
		return "", fmt.Errorf("invalid token_id %q", tokenID)
	}

	rawOrder := &auth.RawOrder{
		Salt:          salt,
		Maker:         common.HexToAddress(l1.Address()),
		Signer:        common.HexToAddress(l1.Address()),
		Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
		TokenID:       tokenIDInt,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Expiration:    big.NewInt(0),
		Nonce:         big.NewInt(0),
		FeeRateBps:    big.NewInt(0),
		Side:          auth.Buy,
		SignatureType: auth.EOA, // Revert to EOA (0)
	}

	sig, err := signer.Sign(rawOrder)
	if err != nil {
		return "", err
	}

	if builderAPIKey != "" {
		log.Info().Str("builder_key_prefix", builderAPIKey[:min(4, len(builderAPIKey))]+"***").
			Msg("builder: attaching builderApiKey to order request")
	} else {
		log.Warn().Msg("builder: no builderApiKey — order will be placed WITHOUT attribution")
	}

	req := &clob.CreateOrderRequest{
		Order: clob.SignedOrder{
			Salt:          salt.String(),
			Maker:         l1.Address(),
			Signer:        l1.Address(),
			Taker:         "0x0000000000000000000000000000000000000000",
			TokenID:       tokenID,
			MakerAmount:   makerAmount.String(),
			TakerAmount:   takerAmount.String(),
			Expiration:    "0",
			Nonce:         "0",
			FeeRateBps:    "0",
			Side:          0, // BUY
			SignatureType: 0, // EOA
			Signature:     sig,
		},
		Owner:         creds.APIKey,
		OrderType:     clob.OrderTypeGTC,
		BuilderApiKey: builderAPIKey,
	}

	resp, err := clobClient.CreateOrder(req)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf("order rejected: %s", resp.ErrorMsg)
	}
	log.Info().
		Str("order_id", resp.OrderID).
		Bool("builder_key_set", builderAPIKey != "").
		Msg("✅ Order accepted by Polymarket")
	return resp.OrderID, nil
}


func ptr[T any](v T) *T { return &v }
