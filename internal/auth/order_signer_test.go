package auth_test

import (
	"math/big"
	"testing"

	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/ethereum/go-ethereum/common"
)

func TestOrderSignerSign(t *testing.T) {
	// Тестовый приватный ключ Hardhat #0 (не использовать в продакшене!)
	l1, err := auth.NewL1Signer("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		t.Fatalf("NewL1Signer: %v", err)
	}

	signer := auth.NewOrderSigner(l1, 137, false)

	order := &auth.RawOrder{
		Salt:          big.NewInt(12345),
		Maker:         common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"),
		Signer:        common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"),
		Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
		TokenID:       big.NewInt(1234567890),
		MakerAmount:   big.NewInt(650000),  // 0.65 USDC
		TakerAmount:   big.NewInt(1000000), // 1.0 share
		Expiration:    big.NewInt(0),
		Nonce:         big.NewInt(0),
		FeeRateBps:    big.NewInt(0),
		Side:          auth.Buy,
		SignatureType: auth.EOA,
	}

	sig, err := signer.Sign(order)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// Подпись: 0x + 130 hex символов (65 байт)
	if len(sig) != 132 {
		t.Errorf("expected sig length 132, got %d: %s", len(sig), sig)
	}
	if sig[:2] != "0x" {
		t.Errorf("expected sig to start with 0x, got: %s", sig[:2])
	}

	// Детерминированность: одинаковый ввод → одинаковая подпись
	sig2, err := signer.Sign(order)
	if err != nil {
		t.Fatal(err)
	}
	if sig != sig2 {
		t.Errorf("expected deterministic signature, got different results")
	}

	t.Logf("signature: %s", sig)
}

func TestRandomSalt(t *testing.T) {
	s1, err := auth.RandomSalt()
	if err != nil {
		t.Fatalf("RandomSalt: %v", err)
	}
	s2, err := auth.RandomSalt()
	if err != nil {
		t.Fatal(err)
	}
	if s1.Cmp(s2) == 0 {
		t.Error("expected different random salts")
	}
	max128 := new(big.Int).Lsh(big.NewInt(1), 128)
	if s1.Cmp(max128) >= 0 {
		t.Errorf("salt exceeds 128 bits")
	}
}
