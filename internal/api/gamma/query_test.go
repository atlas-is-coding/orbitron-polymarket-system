package gamma

import (
	"strings"
	"testing"
)

func TestBuildMarketsQuery_Order(t *testing.T) {
	q := buildMarketsQuery(MarketsParams{
		Order:     "volume_24hr",
		Ascending: false,
		Limit:     100,
	})
	if !strings.Contains(q, "order=volume_24hr") {
		t.Errorf("expected order=volume_24hr in %q", q)
	}
	if !strings.Contains(q, "ascending=false") {
		t.Errorf("expected ascending=false in %q", q)
	}
	if !strings.Contains(q, "limit=100") {
		t.Errorf("expected limit=100 in %q", q)
	}
}

func TestBuildMarketsQuery_Closed(t *testing.T) {
	f := false
	q := buildMarketsQuery(MarketsParams{Closed: &f})
	if !strings.Contains(q, "closed=false") {
		t.Errorf("expected closed=false in %q", q)
	}
}

func TestBuildMarketsQuery_AscendingTrue(t *testing.T) {
	q := buildMarketsQuery(MarketsParams{Order: "liquidity", Ascending: true})
	if !strings.Contains(q, "ascending=true") {
		t.Errorf("expected ascending=true in %q", q)
	}
}

func TestBuildEventsQuery_Order(t *testing.T) {
	q := buildEventsQuery(EventsParams{
		Order:     "volume_24hr",
		Ascending: false,
		Limit:     50,
	})
	if !strings.Contains(q, "order=volume_24hr") {
		t.Errorf("expected order=volume_24hr in %q", q)
	}
}
