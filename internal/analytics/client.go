package analytics

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

// Version is set at build time via -ldflags "-X github.com/atlasdev/orbitron/internal/analytics.Version=v1.2.3"
var Version = "dev"

type TradeReport struct {
	ID        string  `json:"id"`
	MarketID  string  `json:"marketId"`
	AssetID   string  `json:"assetId"`
	Side      string  `json:"side"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Volume    float64 `json:"volume"`
	Strategy  string  `json:"strategy"`
	Timestamp int64   `json:"timestamp"`
}

type Payload struct {
	Address   string        `json:"address"`
	Label     string        `json:"label"`
	Timestamp int64         `json:"timestamp"`
	// Nonce is a random 16-byte hex string. Including it in the canonical hash
	// makes every payload's hash unique, preventing replay attacks.
	Nonce     string        `json:"nonce"`
	// SeqNum is a monotonically increasing counter initialized from
	// time.Now().UnixMilli() at client startup. The server rejects any payload
	// with seqNum <= the last accepted seqNum for this address.
	SeqNum      uint64        `json:"seqNum"`
	Trades      []TradeReport `json:"trades"`
	// PayloadHash is the SHA-256 of the canonical payload (all fields above, in this order).
	PayloadHash string `json:"payloadHash"`
}

// wirePayload is the full body sent to the server (payload + auth + metadata).
type wirePayload struct {
	*Payload
	Signature  string `json:"signature"`
	BotVersion string `json:"botVersion"`
	ChainID    int64  `json:"chainId"`
}

type Client struct {
	httpClient *fasthttp.Client
	signer     *auth.L1Signer
	address    string
	label      string
	endpoint   string
	chainID    int64
	logger     zerolog.Logger
	seqNum     atomic.Uint64
}

func NewClient(signer *auth.L1Signer, label, endpoint string, chainID int64, log zerolog.Logger) *Client {
	c := &Client{
		httpClient: &fasthttp.Client{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		signer:   signer,
		address:  signer.Address(),
		label:    label,
		endpoint: endpoint,
		chainID:  chainID,
		logger:   log.With().Str("component", "analytics-client").Logger(),
	}
	// Initialize from millisecond timestamp so seqNum is always increasing
	// across process restarts (UnixMilli ~1.75e12, safely within JS float64 precision).
	c.seqNum.Store(uint64(time.Now().UnixMilli()))
	return c
}

func (c *Client) preparePayload(trades []TradeReport) (*Payload, string, error) {
	timestamp := time.Now().Unix()
	seqNum := c.seqNum.Add(1)

	var nonceBytes [16]byte
	if _, err := rand.Read(nonceBytes[:]); err != nil {
		return nil, "", fmt.Errorf("generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes[:])

	// Canonical fields MUST be in this exact order to match the server's
	// verifyPayloadIntegrity function.
	type canonicalPayload struct {
		Address   string        `json:"address"`
		Label     string        `json:"label"`
		Timestamp int64         `json:"timestamp"`
		Nonce     string        `json:"nonce"`
		SeqNum    uint64        `json:"seqNum"`
		Trades    []TradeReport `json:"trades"`
	}
	canonical := canonicalPayload{
		Address:   c.address,
		Label:     c.label,
		Timestamp: timestamp,
		Nonce:     nonce,
		SeqNum:    seqNum,
		Trades:    trades,
	}
	canonicalJSON, err := json.Marshal(canonical)
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256(canonicalJSON)
	payloadHash := "0x" + hex.EncodeToString(hash[:])

	p := &Payload{
		Address:     c.address,
		Label:       c.label,
		Timestamp:   timestamp,
		Nonce:       nonce,
		SeqNum:      seqNum,
		Trades:      trades,
		PayloadHash: payloadHash,
	}

	message := fmt.Sprintf("Report Analytics: %s at %d", p.PayloadHash, p.Timestamp)
	signature, err := c.signer.Sign([]byte(message))
	if err != nil {
		return nil, "", err
	}

	return p, signature, nil
}

func (c *Client) Report(ctx context.Context, trades []TradeReport) error {
	if len(trades) == 0 {
		return nil
	}

	payload, signature, err := c.preparePayload(trades)
	if err != nil {
		return fmt.Errorf("prepare payload: %w", err)
	}

	body, err := json.Marshal(&wirePayload{
		Payload:    payload,
		Signature:  signature,
		BotVersion: Version,
		ChainID:    c.chainID,
	})
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.endpoint)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(body)

	var lastErr error
	for i := 0; i < 3; i++ {
		if err := c.httpClient.Do(req, resp); err != nil {
			lastErr = err
		} else if resp.StatusCode() != fasthttp.StatusCreated && resp.StatusCode() != fasthttp.StatusOK {
			lastErr = fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.Body())
		} else {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}

	return fmt.Errorf("failed after 3 attempts: %w", lastErr)
}
