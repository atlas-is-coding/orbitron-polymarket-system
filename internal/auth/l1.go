package auth

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// L1Signer подписывает сообщения приватным ключом Ethereum (L1-аутентификация).
// Используется при создании/отзыве API-ключей через CLOB API.
type L1Signer struct {
	privateKey *ecdsa.PrivateKey
	address    string
}

// NewL1Signer создаёт L1Signer из hex-строки приватного ключа (без 0x).
func NewL1Signer(hexKey string) (*L1Signer, error) {
	hexKey = strings.TrimPrefix(hexKey, "0x")
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("l1: decode private key: %w", err)
	}

	pk, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("l1: parse private key: %w", err)
	}

	addr := crypto.PubkeyToAddress(pk.PublicKey).Hex()
	return &L1Signer{privateKey: pk, address: addr}, nil
}

// Address возвращает Ethereum-адрес, соответствующий приватному ключу.
func (s *L1Signer) Address() string {
	return s.address
}

// Sign подписывает произвольные байты (хэш) и возвращает подпись в формате hex.
// Polymarket ожидает personal_sign (Ethereum prefix + keccak256).
func (s *L1Signer) Sign(data []byte) (string, error) {
	// Ethereum personal_sign: "\x19Ethereum Signed Message:\n" + len + data
	hash := crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data))),
		data,
	)

	sig, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return "", fmt.Errorf("l1: sign: %w", err)
	}

	// go-ethereum возвращает [R || S || V], V ∈ {0,1}
	// Ethereum ожидает V ∈ {27,28}
	sig[64] += 27

	return "0x" + hex.EncodeToString(sig), nil
}

// L1Headers возвращает заголовки для L1-аутентификации (для /auth/api-key endpoint).
func (s *L1Signer) L1Headers(timestamp, nonce string) (map[string]string, error) {
	// Polymarket L1 подписывает: timestamp + ":" + nonce
	msg := []byte(timestamp + ":" + nonce)
	sig, err := s.Sign(msg)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"POLY_ADDRESS":   s.address,
		"POLY_TIMESTAMP": timestamp,
		"POLY_NONCE":     nonce,
		"POLY_SIGNATURE": sig,
	}, nil
}
