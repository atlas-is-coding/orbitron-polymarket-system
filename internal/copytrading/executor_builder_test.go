package copytrading

import "testing"

func TestWithBuilderKey_SetsField(t *testing.T) {
	e := &OrderExecutor{}
	result := e.WithBuilderKey("builder-key-123")
	if result != e {
		t.Fatal("WithBuilderKey must return the same executor instance")
	}
	if e.builderAPIKey != "builder-key-123" {
		t.Fatalf("builderAPIKey = %q, want builder-key-123", e.builderAPIKey)
	}
}

func TestWithBuilderKey_Empty(t *testing.T) {
	e := &OrderExecutor{builderAPIKey: "old"}
	e.WithBuilderKey("")
	if e.builderAPIKey != "" {
		t.Fatalf("expected empty builderAPIKey after WithBuilderKey('')")
	}
}
