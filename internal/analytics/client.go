package analytics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/atlasdev/orbitron/internal/auth"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type TradeReport struct {
	ID           string  `json:"id"`
	MarketID     string  `json:"marketId"`
	AssetID      string  `json:"assetId"`
	Side         string  `json:"side"`
	Price        float64 `json:"price"`
	Size         float64 `json:"size"`
	Volume       float64 `json:"volume"`
	Strategy     string  `json:"strategy"`
	Timestamp    int64   `json:"timestamp"`
}

type Payload struct {
	Address     string        `json:"address"`
	Label       string        `json:"label"`
	Timestamp   int64         `json:"timestamp"`
	Trades      []TradeReport `json:"trades"`
	PayloadHash string        `json:"payloadHash"`
}

type Client struct {
	httpClient *fasthttp.Client
	signer     *auth.L1Signer
	address    string
	label      string
	endpoint   string
	logger     zerolog.Logger
}

func NewClient(signer *auth.L1Signer, label, endpoint string, log zerolog.Logger) *Client {
	return &Client{
		httpClient: &fasthttp.Client{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		signer:   signer,
		address:  signer.Address(),
		label:    label,
		endpoint: endpoint,
		logger:   log.With().Str("component", "analytics-client").Logger(),
	}
}

func (c *Client) preparePayload(trades []TradeReport) (*Payload, string, error) {
	timestamp := time.Now().Unix()
	
	p := &Payload{
		Address:   c.address,
		Label:     c.label,
		Timestamp: timestamp,
		Trades:    trades,
	}

	basicJSON, err := json.Marshal(p)
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256(basicJSON)
	p.PayloadHash = "0x" + hex.EncodeToString(hash[:])

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

	body, err := json.Marshal(struct {
		*Payload
		Signature string `json:"signature"`
	}{
		Payload:   payload,
		Signature: signature,
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
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		if resp.StatusCode() != fasthttp.StatusCreated && resp.StatusCode() != fasthttp.StatusOK {
			lastErr = fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.Body())
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed after 3 attempts: %w", lastErr)
}
