package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestL2Headers_Structure(t *testing.T) {
	creds := &L2Credentials{
		APIKey:     "test-api-key",
		APISecret:  "dGVzdC1zZWNyZXQ=",
		Passphrase: "test-pass",
		Address:    "0x1234",
	}
	headers, err := creds.L2Headers("GET", "/orders", "")
	require.NoError(t, err)
	assert.Equal(t, creds.Address, headers["POLY_ADDRESS"])
	assert.Equal(t, creds.APIKey, headers["POLY_API_KEY"])
	assert.Equal(t, creds.Passphrase, headers["POLY_PASSPHRASE"])
	assert.NotEmpty(t, headers["POLY_TIMESTAMP"])
	assert.NotEmpty(t, headers["POLY_SIGNATURE"])
}

func TestL2Headers_DifferentMethods(t *testing.T) {
	creds := &L2Credentials{
		APIKey:     "key",
		APISecret:  "c2VjcmV0",
		Passphrase: "pass",
		Address:    "0xabc",
	}
	h1, err := creds.L2Headers("GET", "/orders", "")
	require.NoError(t, err)
	h2, err := creds.L2Headers("POST", "/order", `{"test":"body"}`)
	require.NoError(t, err)
	assert.NotEqual(t, h1["POLY_SIGNATURE"], h2["POLY_SIGNATURE"])
}
