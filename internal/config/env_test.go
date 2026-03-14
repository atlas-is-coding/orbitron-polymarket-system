package config

import "testing"

func TestEnvOverlay(t *testing.T) {
	t.Setenv("POLY_PRIVATE_KEY", "env-pk")
	t.Setenv("TELEGRAM_BOT_TOKEN", "env-tg")
	t.Setenv("POLY_API_KEY", "env-ak")
	t.Setenv("POLY_API_SECRET", "env-as")
	t.Setenv("POLY_PASSPHRASE", "env-pp")
	t.Setenv("WEBUI_JWT_SECRET", "env-jwt")
	t.Setenv("POLY_PROXY_USERNAME", "env-pu")
	t.Setenv("POLY_PROXY_PASSWORD", "env-pw")

	cfg := &Config{}
	cfg.applyEnvOverlay()

	if cfg.Wallets[0].PrivateKey != "env-pk" {
		t.Fatalf("private key: want env-pk, got %q", cfg.Wallets[0].PrivateKey)
	}
	if cfg.Telegram.BotToken != "env-tg" {
		t.Fatalf("telegram token not overridden")
	}
	if cfg.Wallets[0].APIKey != "env-ak" {
		t.Fatalf("api key not overridden")
	}
	if cfg.WebUI.JWTSecret != "env-jwt" {
		t.Fatalf("jwt secret not overridden")
	}
	if cfg.Proxy.Username != "env-pu" {
		t.Fatalf("proxy username not overridden")
	}
	if cfg.Proxy.Password != "env-pw" {
		t.Fatalf("proxy password not overridden")
	}
}
