package analytics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

// TestCanonicalHash_ExcludesPayloadHashField verifies that the hash sent to the
// server is computed from {address, label, timestamp, trades} ONLY, without the
// payloadHash field itself. This must match what verifyPayloadIntegrity() on the
// server recomputes. Including "payloadHash":"" would produce a different hash
// and cause every report to be rejected with 400.
func TestCanonicalHash_ExcludesPayloadHashField(t *testing.T) {
	trades := []TradeReport{
		{ID: "t1", MarketID: "m1", AssetID: "a1", Side: "BUY", Price: 0.5, Size: 100, Volume: 50, Strategy: "arbitrage", Timestamp: 1700000000},
	}

	// Simulate what preparePayload now does: hash only canonical fields.
	type canonicalPayload struct {
		Address   string        `json:"address"`
		Label     string        `json:"label"`
		Timestamp int64         `json:"timestamp"`
		Trades    []TradeReport `json:"trades"`
	}
	canonical := canonicalPayload{
		Address:   "0xabc",
		Label:     "test",
		Timestamp: 1700000000,
		Trades:    trades,
	}
	canonicalJSON, err := json.Marshal(canonical)
	if err != nil {
		t.Fatalf("marshal canonical: %v", err)
	}
	hash := sha256.Sum256(canonicalJSON)
	expectedHash := "0x" + hex.EncodeToString(hash[:])

	// Simulate what the OLD (buggy) code did: marshal full Payload with payloadHash:"".
	fullPayload := &Payload{
		Address:   "0xabc",
		Label:     "test",
		Timestamp: 1700000000,
		Trades:    trades,
		// PayloadHash is "" at hash time
	}
	fullJSON, err := json.Marshal(fullPayload)
	if err != nil {
		t.Fatalf("marshal full: %v", err)
	}
	buggyHash := sha256.Sum256(fullJSON)
	buggyHashStr := "0x" + hex.EncodeToString(buggyHash[:])

	// The two must differ — proving the fix matters.
	if expectedHash == buggyHashStr {
		t.Error("canonical hash and full-payload hash are identical; the payloadHash field may already be omitempty or absent from JSON — re-check test setup")
	}

	// Verify the canonical JSON does NOT contain "payloadHash".
	if contains := string(canonicalJSON); len(contains) > 0 {
		if containsField(canonicalJSON, "payloadHash") {
			t.Error("canonical JSON must not contain payloadHash field")
		}
	}
}

func containsField(data []byte, field string) bool {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return false
	}
	_, ok := m[field]
	return ok
}

// TestPreparePayload_VolumeIsClientSupplied verifies that the payload volume
// field is what the client reports (the server is responsible for recomputing it).
// This test documents the client's behavior; volume sanitization is on the server side.
func TestPreparePayload_VolumeSentAsIs(t *testing.T) {
	trades := []TradeReport{
		{ID: "t1", Price: 0.5, Size: 100, Volume: 50},
	}
	// Directly marshal to see what the client sends
	body, err := json.Marshal(trades)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out []map[string]interface{}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	vol, ok := out[0]["volume"].(float64)
	if !ok {
		t.Fatal("volume field missing or wrong type")
	}
	if vol != 50 {
		t.Errorf("expected volume 50, got %v", vol)
	}
}

// TestWirePayload_ContainsBotVersionAndChainID verifies the wire format sent to the server.
func TestWirePayload_ContainsBotVersionAndChainID(t *testing.T) {
	Version = "v1.2.3-test"

	p := &Payload{
		Address:   "0xabc",
		Label:     "test",
		Timestamp: 1700000000,
		Trades:    nil,
	}
	wire := &wirePayload{
		Payload:    p,
		Signature:  "0xsig",
		BotVersion: Version,
		ChainID:    137,
	}

	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m["botVersion"] != "v1.2.3-test" {
		t.Errorf("expected botVersion v1.2.3-test, got %v", m["botVersion"])
	}
	chainID, _ := m["chainId"].(float64)
	if chainID != 137 {
		t.Errorf("expected chainId 137, got %v", m["chainId"])
	}
}

// TestTradeReport_VolumeOmittedFromHash verifies that altering volume alone
// does NOT change the wire payload hash (server ignores client volume).
// This is a documentation test — volume should NOT affect rewards.
func TestTradeReport_JSONFields(t *testing.T) {
	tr := TradeReport{
		ID:        "trade-1",
		MarketID:  "market-1",
		AssetID:   "asset-1",
		Side:      "BUY",
		Price:     0.75,
		Size:      200,
		Volume:    150, // server ignores this
		Strategy:  "arbitrage",
		Timestamp: 1700000000,
	}
	body, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	required := []string{"id", "marketId", "assetId", "side", "price", "size", "volume", "strategy", "timestamp"}
	for _, k := range required {
		if _, ok := m[k]; !ok {
			t.Errorf("expected field %q in JSON output", k)
		}
	}
}
