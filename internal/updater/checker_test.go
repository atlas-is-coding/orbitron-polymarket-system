package updater_test

import (
	"context"
	"testing"

	"github.com/atlasdev/orbitron/internal/updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteIsNewer(t *testing.T) {
	cases := []struct {
		remote, local string
		wantNewer     bool
	}{
		{"1.1.0", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.1.0", false},
		{"1.10.0", "1.9.0", true},  // 10 > 9
		{"2.0.0", "1.99.99", true},
		{"bad", "1.0.0", false},    // malformed remote → not newer
	}
	for _, c := range cases {
		got := updater.RemoteIsNewer(c.remote, c.local)
		assert.Equal(t, c.wantNewer, got, "remote=%s local=%s", c.remote, c.local)
	}
}

func TestScheduleUpdate_NotIdle_SavesPending(t *testing.T) {
	dir := t.TempDir()
	p := updater.NewPending(dir)
	notIdle := func() bool { return false }

	updater.ScheduleUpdateWith(context.Background(), notIdle, p, "1.1.0", "https://example.com/bin", nil)

	got, ok := p.Load()
	require.True(t, ok)
	assert.Equal(t, "1.1.0", got.Version)
}
