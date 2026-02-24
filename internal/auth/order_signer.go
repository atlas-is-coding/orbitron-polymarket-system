package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Адреса CTF Exchange контрактов на Polygon (chainId=137)
const (
	CTFExchangeMain    = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"
	CTFExchangeNegRisk = "0xC5d563A36AE78145C45a50134d48A1215220f80a"
)

// OrderSide — сторона ордера в EIP-712 (0=BUY, 1=SELL)
type OrderSide int

const (
	Buy  OrderSide = 0
	Sell OrderSide = 1
)

// SignatureType — тип подписи (0=EOA, 1=POLY_PROXY, 2=POLY_GNOSIS_SAFE)
type SignatureType int

const (
	EOA        SignatureType = 0
	PolyProxy  SignatureType = 1
	GnosisSafe SignatureType = 2
)

// RawOrder — неподписанный ордер для хеширования по EIP-712.
type RawOrder struct {
	Salt          *big.Int
	Maker         common.Address
	Signer        common.Address
	Taker         common.Address
	TokenID       *big.Int
	MakerAmount   *big.Int
	TakerAmount   *big.Int
	Expiration    *big.Int
	Nonce         *big.Int
	FeeRateBps    *big.Int
	Side          OrderSide
	SignatureType SignatureType
}

// ORDER_TYPEHASH — keccak256 хеш строки типа ордера EIP-712.
var ORDER_TYPEHASH = crypto.Keccak256Hash([]byte(
	"Order(uint256 salt,address maker,address signer,address taker,uint256 tokenId," +
		"uint256 makerAmount,uint256 takerAmount,uint256 expiration,uint256 nonce," +
		"uint256 feeRateBps,uint8 side,uint8 signatureType)",
))

// OrderSigner подписывает ордера приватным ключом Ethereum (EIP-712).
type OrderSigner struct {
	l1        *L1Signer
	domainSep [32]byte
}

// NewOrderSigner создаёт OrderSigner для указанного exchange контракта.
// negRisk=true использует NegRisk Exchange адрес.
func NewOrderSigner(l1 *L1Signer, chainID int64, negRisk bool) *OrderSigner {
	addr := CTFExchangeMain
	if negRisk {
		addr = CTFExchangeNegRisk
	}
	exchangeAddr := common.HexToAddress(addr)

	// EIP-712 domain separator
	domainTypeHash := crypto.Keccak256Hash([]byte(
		"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)",
	))
	nameHash := crypto.Keccak256Hash([]byte("Polymarket CTF Exchange"))
	versionHash := crypto.Keccak256Hash([]byte("1"))

	chainIDPadded := padBigInt(big.NewInt(chainID))
	addrPadded := padAddress(exchangeAddr)

	domainSepBytes := crypto.Keccak256(
		domainTypeHash.Bytes(),
		nameHash.Bytes(),
		versionHash.Bytes(),
		chainIDPadded,
		addrPadded,
	)
	var domainSep [32]byte
	copy(domainSep[:], domainSepBytes)

	return &OrderSigner{
		l1:        l1,
		domainSep: domainSep,
	}
}

// Sign вычисляет EIP-712 хеш ордера и подписывает его приватным ключом.
// Возвращает hex-строку подписи с префиксом "0x".
func (s *OrderSigner) Sign(order *RawOrder) (string, error) {
	orderHash := s.hashOrder(order)

	// EIP-712 финальный хеш: keccak256("\x19\x01" + domainSeparator + structHash)
	finalHash := crypto.Keccak256(
		[]byte("\x19\x01"),
		s.domainSep[:],
		orderHash[:],
	)

	sig, err := crypto.Sign(finalHash, s.l1.privateKey)
	if err != nil {
		return "", fmt.Errorf("order signer: sign: %w", err)
	}
	// go-ethereum возвращает [R || S || V], V ∈ {0,1}; Ethereum ожидает V ∈ {27,28}
	sig[64] += 27
	return "0x" + hex.EncodeToString(sig), nil
}

// hashOrder вычисляет keccak256 хеш структуры ордера по EIP-712.
func (s *OrderSigner) hashOrder(o *RawOrder) [32]byte {
	encoded := make([]byte, 0, 32*13)
	encoded = append(encoded, ORDER_TYPEHASH.Bytes()...)
	encoded = append(encoded, padBigInt(o.Salt)...)
	encoded = append(encoded, padAddress(o.Maker)...)
	encoded = append(encoded, padAddress(o.Signer)...)
	encoded = append(encoded, padAddress(o.Taker)...)
	encoded = append(encoded, padBigInt(o.TokenID)...)
	encoded = append(encoded, padBigInt(o.MakerAmount)...)
	encoded = append(encoded, padBigInt(o.TakerAmount)...)
	encoded = append(encoded, padBigInt(o.Expiration)...)
	encoded = append(encoded, padBigInt(o.Nonce)...)
	encoded = append(encoded, padBigInt(o.FeeRateBps)...)
	encoded = append(encoded, padUint8(uint8(o.Side))...)
	encoded = append(encoded, padUint8(uint8(o.SignatureType))...)
	return crypto.Keccak256Hash(encoded)
}

// RandomSalt генерирует случайный 128-битный salt для ордера.
func RandomSalt() (*big.Int, error) {
	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("random salt: %w", err)
	}
	return n, nil
}

// --- Вспомогательные функции ABI-кодирования ---

func padBigInt(n *big.Int) []byte {
	if n == nil {
		return make([]byte, 32)
	}
	b := n.Bytes()
	pad := make([]byte, 32)
	copy(pad[32-len(b):], b)
	return pad
}

func padAddress(addr common.Address) []byte {
	pad := make([]byte, 32)
	copy(pad[12:], addr.Bytes())
	return pad
}

func padUint8(v uint8) []byte {
	pad := make([]byte, 32)
	pad[31] = v
	return pad
}
