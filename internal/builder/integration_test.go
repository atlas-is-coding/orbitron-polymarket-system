package builder_test

import (
	"encoding/json"
	"testing"

	"github.com/atlasdev/orbitron/internal/api/clob"
)

// TestCreateOrderRequest_BuilderApiKey_OmitemptyJSON verifies that:
// - builderApiKey field is present in JSON when set
// - builderApiKey field is absent in JSON when empty (omitempty)
func TestCreateOrderRequest_BuilderApiKey_OmitemptyJSON(t *testing.T) {
	t.Run("present when set", func(t *testing.T) {
		req := clob.CreateOrderRequest{
			Owner:         "0xabc",
			BuilderApiKey: "test-key-123",
		}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, ok := m["builderApiKey"]; !ok {
			t.Errorf("builderApiKey missing from JSON when key is set; got: %s", string(b))
		}
		if m["builderApiKey"] != "test-key-123" {
			t.Errorf("builderApiKey value mismatch: got %v", m["builderApiKey"])
		}
	})

	t.Run("absent when empty (omitempty)", func(t *testing.T) {
		req := clob.CreateOrderRequest{
			Owner:         "0xabc",
			BuilderApiKey: "",
		}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, ok := m["builderApiKey"]; ok {
			t.Errorf("builderApiKey should be absent from JSON when empty; got: %s", string(b))
		}
	})
}
