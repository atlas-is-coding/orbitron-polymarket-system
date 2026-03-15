package testutil

import (
	"os"
	"strings"
	"testing"

	"github.com/atlasdev/orbitron/internal/api"
	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/api/data"
	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/auth"
)

const (
	ClobURL  = "https://clob.polymarket.com"
	GammaURL = "https://gamma-api.polymarket.com"
	DataURL  = "https://data-api.polymarket.com"
)

// NewCLOBClient возвращает публичный CLOB клиент (без auth).
func NewCLOBClient() *clob.Client {
	h := api.NewClient(ClobURL, 10, 2)
	return clob.NewClient(h, nil)
}

// NewGammaClient возвращает Gamma API клиент.
func NewGammaClient() *gamma.Client {
	h := api.NewClient(GammaURL, 10, 2)
	return gamma.NewClient(h)
}

// NewDataClient возвращает Data API клиент.
func NewDataClient() *data.Client {
	h := api.NewClient(DataURL, 10, 2)
	return data.NewClient(h)
}

// LoadPrivateKey читает POLY_PRIVATE_KEY из env, обрезает "0x", вызывает t.Skip если не задан.
func LoadPrivateKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("POLY_PRIVATE_KEY")
	if key == "" {
		t.Skip("POLY_PRIVATE_KEY not set — skipping integration test")
	}
	return strings.TrimPrefix(key, "0x")
}

// LoadL1Signer создаёт L1Signer из POLY_PRIVATE_KEY, пропускает тест если ключ не задан.
func LoadL1Signer(t *testing.T) *auth.L1Signer {
	t.Helper()
	rawKey := LoadPrivateKey(t)
	l1, err := auth.NewL1Signer(rawKey)
	if err != nil {
		t.Fatalf("testutil: NewL1Signer: %v", err)
	}
	return l1
}

// LoadL2Creds выводит L2 credentials через DeriveAPIKey. Вызывает t.Skip если нет ключа.
func LoadL2Creds(t *testing.T) (*auth.L1Signer, *auth.L2Credentials) {
	t.Helper()
	l1 := LoadL1Signer(t)
	pubClient := NewCLOBClient()
	creds, err := pubClient.DeriveAPIKey(l1)
	if err != nil {
		t.Fatalf("testutil: DeriveAPIKey: %v", err)
	}
	return l1, creds
}

// NewAuthCLOBClient возвращает CLOB клиент с L2 credentials.
func NewAuthCLOBClient(creds *auth.L2Credentials) *clob.Client {
	h := api.NewClient(ClobURL, 10, 2)
	return clob.NewClient(h, creds)
}
