package updater_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/atlasdev/orbitron/internal/updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPending_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := updater.NewPending(dir)

	_, ok := p.Load()
	assert.False(t, ok, "no file → Load returns false")

	p.Save("1.1.0", "https://example.com/bin")
	got, ok := p.Load()
	require.True(t, ok)
	assert.Equal(t, "1.1.0", got.Version)
	assert.Equal(t, "https://example.com/bin", got.BinaryURL)

	p.Clear()
	_, ok = p.Load()
	assert.False(t, ok, "after Clear → Load returns false")
}

func TestPending_MalformedFile(t *testing.T) {
	dir := t.TempDir()
	p := updater.NewPending(dir)

	err := os.WriteFile(filepath.Join(dir, ".update_pending"), []byte("not-json"), 0o600)
	require.NoError(t, err)

	_, ok := p.Load()
	assert.False(t, ok, "malformed JSON → Load returns false and deletes file")

	// File must be deleted after malformed read
	_, err = os.Stat(filepath.Join(dir, ".update_pending"))
	assert.True(t, os.IsNotExist(err))
}
