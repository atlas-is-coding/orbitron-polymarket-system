package license

import "testing"

func TestTokenRoundtrip(t *testing.T) {
	plain := "my-secret-app-token"
	encoded := encodeToken(plain)
	if encoded == plain {
		t.Fatal("encoded must differ from plain")
	}
	got := decodeToken(encoded)
	if got != plain {
		t.Fatalf("want %q, got %q", plain, got)
	}
}

func TestEmptyTokenReturnsEmpty(t *testing.T) {
	if decodeToken("") != "" {
		t.Fatal("empty encoded must return empty")
	}
}
