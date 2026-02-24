// Package auth реализует двухуровневую аутентификацию Polymarket.
// L1 — EIP-712 подпись приватным ключом Ethereum (для создания API ключей).
// L2 — HMAC-SHA256 подпись с API key/secret/passphrase (для торговых операций).
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

// L2Credentials содержит учётные данные для L2-аутентификации.
type L2Credentials struct {
	APIKey     string
	APISecret  string
	Passphrase string
	Address    string
}

// L2Headers возвращает заголовки для L2-аутентификации.
// method — HTTP-метод (GET, POST, DELETE), path — путь запроса, body — тело запроса (или пустая строка).
func (c *L2Credentials) L2Headers(method, path, body string) (map[string]string, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	msg := ts + method + path + body

	sig, err := hmacSHA256Base64(c.APISecret, msg)
	if err != nil {
		return nil, fmt.Errorf("l2 sign: %w", err)
	}

	return map[string]string{
		"POLY_ADDRESS":    c.Address,
		"POLY_TIMESTAMP":  ts,
		"POLY_API_KEY":    c.APIKey,
		"POLY_PASSPHRASE": c.Passphrase,
		"POLY_SIGNATURE":  sig,
	}, nil
}

func hmacSHA256Base64(secret, message string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(message)); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
