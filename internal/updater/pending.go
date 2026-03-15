package updater

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const pendingFile = ".update_pending"

// PendingUpdate holds the deferred update state written to disk.
type PendingUpdate struct {
	Version   string `json:"version"`
	BinaryURL string `json:"binary_url"`
}

// Pending manages the .update_pending file in a given directory.
type Pending struct {
	path string
}

// NewPending returns a Pending rooted at dir.
func NewPending(dir string) *Pending {
	return &Pending{path: filepath.Join(dir, pendingFile)}
}

// Save writes version and binaryURL to the pending file.
func (p *Pending) Save(version, binaryURL string) {
	data, _ := json.Marshal(PendingUpdate{Version: version, BinaryURL: binaryURL})
	if err := os.WriteFile(p.path, data, 0o600); err != nil {
		log.Warn().Err(err).Msg("updater: failed to save pending update")
	}
}

// Load reads the pending file. Returns (nil, false) when absent or malformed.
func (p *Pending) Load() (*PendingUpdate, bool) {
	data, err := os.ReadFile(p.path)
	if os.IsNotExist(err) {
		return nil, false
	}
	var u PendingUpdate
	if err := json.Unmarshal(data, &u); err != nil || u.Version == "" {
		log.Warn().Msg("updater: malformed .update_pending — deleting")
		_ = os.Remove(p.path)
		return nil, false
	}
	return &u, true
}

// Clear deletes the pending file.
func (p *Pending) Clear() {
	_ = os.Remove(p.path)
}
