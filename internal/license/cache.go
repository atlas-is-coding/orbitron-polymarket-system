package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/hkdf"
)

const cacheTTL = 30 * 24 * time.Hour
const hkdfInfo = "polytrade-bot cache v1"

type cacheEnvelope struct {
	Nonce      []byte    `json:"nonce"`
	Ciphertext []byte    `json:"ciphertext"`
	SavedAt    time.Time `json:"saved_at"`
}

func cachePath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "polytrade-bot", "creds.enc"), nil
}

func deriveKey(token string) ([]byte, error) {
	key := make([]byte, 32)
	r := hkdf.New(sha256.New, []byte(token), nil, []byte(hkdfInfo))
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, err
	}
	return key, nil
}

func saveCache(token string, creds *BuilderCredentials) error {
	plain, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	key, err := deriveKey(token)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	env := cacheEnvelope{
		Nonce:      nonce,
		Ciphertext: gcm.Seal(nil, nonce, plain, nil),
		SavedAt:    time.Now(),
	}
	p, err := cachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, _ := json.Marshal(env)
	return os.WriteFile(p, data, 0600)
}

func loadCache(token string) (*BuilderCredentials, error) {
	p, err := cachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var env cacheEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	if time.Since(env.SavedAt) > cacheTTL {
		return nil, errors.New("cache expired")
	}
	key, err := deriveKey(token)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, env.Nonce, env.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	var creds BuilderCredentials
	if err := json.Unmarshal(plain, &creds); err != nil {
		return nil, err
	}
	if !creds.ExpiresAt.IsZero() && time.Now().After(creds.ExpiresAt) {
		return nil, errors.New("cache expired: credentials past ExpiresAt")
	}
	return &creds, nil
}
