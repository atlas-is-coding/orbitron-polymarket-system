package markets_test

import (
	"testing"

	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/markets"
)

func TestServiceAddRemoveAlert(t *testing.T) {
	svc := markets.NewService(nil, nil)

	id := svc.AddAlert(markets.AlertRule{
		ConditionID: "0xabc",
		TokenID:     "123",
		Direction:   "above",
		Threshold:   0.80,
	})
	if id == "" {
		t.Fatal("AddAlert: expected non-empty ID")
	}

	alerts := svc.Alerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].ID != id {
		t.Fatalf("expected ID %s, got %s", id, alerts[0].ID)
	}

	svc.RemoveAlert(id)
	if len(svc.Alerts()) != 0 {
		t.Fatal("expected 0 alerts after remove")
	}
}

func TestServiceGetByTag(t *testing.T) {
	svc := markets.NewService(nil, nil)
	svc.SetMarketsForTest([]gamma.Market{
		{ConditionID: "0x1", Tags: []gamma.Tag{{Slug: "crypto"}}},
		{ConditionID: "0x2", Tags: []gamma.Tag{{Slug: "politics"}}},
		{ConditionID: "0x3", Tags: []gamma.Tag{{Slug: "crypto"}}},
	})

	result := svc.GetByTag("crypto")
	if len(result) != 2 {
		t.Fatalf("GetByTag(crypto): expected 2, got %d", len(result))
	}
}

func TestServiceGetByTagAll(t *testing.T) {
	svc := markets.NewService(nil, nil)
	svc.SetMarketsForTest([]gamma.Market{
		{ConditionID: "0x1"},
		{ConditionID: "0x2"},
	})

	result := svc.GetByTag("")
	if len(result) != 2 {
		t.Fatalf("GetByTag(empty): expected 2, got %d", len(result))
	}
}

func TestServiceGetMarket(t *testing.T) {
	svc := markets.NewService(nil, nil)
	svc.SetMarketsForTest([]gamma.Market{
		{ConditionID: "0xABC", Question: "Will it?"},
	})

	m, ok := svc.GetMarket("0xABC")
	if !ok {
		t.Fatal("GetMarket: expected found")
	}
	if m.Question != "Will it?" {
		t.Fatalf("GetMarket: wrong market returned")
	}

	_, ok = svc.GetMarket("nonexistent")
	if ok {
		t.Fatal("GetMarket nonexistent: expected not found")
	}
}
