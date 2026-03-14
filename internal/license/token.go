package license

import "encoding/hex"

// rawToken is set at build time via:
//   go build -ldflags="-X 'github.com/atlasdev/orbitron/internal/license.rawToken=ENCODED'"
// where ENCODED = hex(realToken XOR xorKey). Generate with: go run ./cmd/tokenenc encode REAL_TOKEN
var rawToken = ""

// xorKey is a fixed key embedded in the binary as a byte array (not a string literal).
// Change this value before your first production build and never change it again.
var xorKey = [16]byte{0x3f, 0xa1, 0x7c, 0x54, 0x9e, 0x2b, 0x61, 0xd8,
	0x05, 0xf3, 0x48, 0xbc, 0x77, 0x1a, 0xe9, 0x30}

// AppToken returns the decoded plaintext app token. Returns "" if rawToken is unset.
func AppToken() string {
	return decodeToken(rawToken)
}

func encodeToken(plain string) string {
	b := xorBytes([]byte(plain))
	return hex.EncodeToString(b)
}

func decodeToken(encoded string) string {
	if encoded == "" {
		return ""
	}
	b, err := hex.DecodeString(encoded)
	if err != nil {
		return ""
	}
	return string(xorBytes(b))
}

func xorBytes(src []byte) []byte {
	out := make([]byte, len(src))
	for i, c := range src {
		out[i] = c ^ xorKey[i%len(xorKey)]
	}
	return out
}

// EncodeTokenPublic encodes a plaintext token for use with go build -ldflags.
func EncodeTokenPublic(plain string) string { return encodeToken(plain) }
