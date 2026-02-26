package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPrivKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewL1Signer_Valid(t *testing.T) {
	l1, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)
	assert.NotNil(t, l1)
	assert.NotEmpty(t, l1.Address())
	assert.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", l1.Address())
}

func TestNewL1Signer_With0xPrefix(t *testing.T) {
	l1WithPrefix, err := NewL1Signer("0x" + testPrivKey)
	require.NoError(t, err)
	l1Without, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)
	assert.Equal(t, l1Without.Address(), l1WithPrefix.Address())
}

func TestNewL1Signer_InvalidKey(t *testing.T) {
	_, err := NewL1Signer("not-hex-key")
	assert.Error(t, err)
}

func TestNewL1Signer_EmptyKey(t *testing.T) {
	_, err := NewL1Signer("")
	assert.Error(t, err)
}

func TestL1Signer_Sign(t *testing.T) {
	l1, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)
	sig, err := l1.Sign([]byte("hello"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(sig, "0x"), "signature must start with 0x")
	assert.Len(t, sig, 132, "65-byte signature = 130 hex chars + '0x' prefix")
}

func TestL1Signer_Sign_Deterministic(t *testing.T) {
	l1, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)
	sig1, err := l1.Sign([]byte("test"))
	require.NoError(t, err)
	sig2, err := l1.Sign([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, sig1, sig2)
}

func TestL1Headers(t *testing.T) {
	l1, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)
	headers, err := l1.L1Headers("1700000000", "0")
	require.NoError(t, err)
	assert.Equal(t, l1.Address(), headers["POLY_ADDRESS"])
	assert.Equal(t, "1700000000", headers["POLY_TIMESTAMP"])
	assert.Equal(t, "0", headers["POLY_NONCE"])
	assert.True(t, strings.HasPrefix(headers["POLY_SIGNATURE"], "0x"))
}
