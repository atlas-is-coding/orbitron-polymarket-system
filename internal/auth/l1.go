package auth

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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
// Polymarket ожидает EIP-712 подпись структуры ClobAuth.
func (s *L1Signer) L1Headers(timestamp, nonce string, chainID int64) (map[string]string, error) {
	sig, err := s.SignAuth(timestamp, nonce, chainID)
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

// SignAuth вычисляет EIP-712 подпись для аутентификации на CLOB API.
func (s *L1Signer) SignAuth(timestamp, nonce string, chainID int64) (string, error) {
	// 1. Domain Separator
	domainTypeHash := crypto.Keccak256Hash([]byte(
		"EIP712Domain(string name,string version,uint256 chainId)",
	))
	nameHash := crypto.Keccak256Hash([]byte("ClobAuthDomain"))
	versionHash := crypto.Keccak256Hash([]byte("1"))

	domainSep := crypto.Keccak256(
		domainTypeHash.Bytes(),
		nameHash.Bytes(),
		versionHash.Bytes(),
		padBigInt(big.NewInt(chainID)),
	)

	// 2. Struct Hash (ClobAuth)
	clobAuthTypeHash := crypto.Keccak256Hash([]byte(
		"ClobAuth(address address,string timestamp,uint256 nonce,string message)",
	))
	
	msg := "This message attests that I control the given wallet"
	msgHash := crypto.Keccak256Hash([]byte(msg))
	tsHash := crypto.Keccak256Hash([]byte(timestamp))
	
	nonceInt := new(big.Int)
	if _, ok := nonceInt.SetString(nonce, 10); !ok {
		return "", fmt.Errorf("invalid nonce %q", nonce)
	}

	encoded := make([]byte, 0, 32*5)
	encoded = append(encoded, clobAuthTypeHash.Bytes()...)
	encoded = append(encoded, padAddress(common.HexToAddress(s.address))...)
	encoded = append(encoded, tsHash.Bytes()...)
	encoded = append(encoded, padBigInt(nonceInt)...)
	encoded = append(encoded, msgHash.Bytes()...)
	structHash := crypto.Keccak256(encoded)

	// 3. Final Hash
	finalHash := crypto.Keccak256(
		[]byte("\x19\x01"),
		domainSep,
		structHash,
	)

	sig, err := crypto.Sign(finalHash, s.privateKey)
	if err != nil {
		return "", fmt.Errorf("l1: sign auth: %w", err)
	}

	sig[64] += 27
	return "0x" + hex.EncodeToString(sig), nil
}
