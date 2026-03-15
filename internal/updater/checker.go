package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/atlasdev/orbitron/internal/license"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
)

// VersionServerURL is the endpoint to check for updates.
// Override in tests or via ldflags.
var VersionServerURL = "https://getorbitron.net/api/v1/version"

type versionResponse struct {
	Version      string            `json:"version"`
	ReleaseNotes string            `json:"release_notes"`
	PublishedAt  string            `json:"published_at"`
	Binaries     map[string]string `json:"binaries"`
}

// RemoteIsNewer returns true when remote version string is strictly greater than local.
// Both strings are plain semver without the "v" prefix (e.g. "1.1.0").
func RemoteIsNewer(remote, local string) bool {
	r := "v" + remote
	l := "v" + local
	if !semver.IsValid(r) || !semver.IsValid(l) {
		return false
	}
	return semver.Compare(r, l) > 0
}

// Start checks for updates on startup (handling any pending update first), then
// rechecks every 6 hours until ctx is cancelled.
func Start(ctx context.Context, isIdle func() bool, n *Notifier, p *Pending) {
	dir := Dir()

	// Handle a deferred update from a previous run.
	if u, ok := p.Load(); ok {
		n.Notify(u.Version, "", "")
		RunUpdate(ctx, u.Version, u.BinaryURL, dir, p, n)
	}

	checkVersion(ctx, dir, isIdle, n, p)

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkVersion(ctx, dir, isIdle, n, p)
		}
	}
}

func checkVersion(ctx context.Context, dir string, isIdle func() bool, n *Notifier, p *Pending) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(VersionServerURL)
	if err != nil {
		log.Warn().Err(err).Msg("updater: version check failed")
		return
	}
	defer resp.Body.Close()

	var vr versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil || vr.Version == "" {
		log.Warn().Msg("updater: invalid version response")
		return
	}

	if !RemoteIsNewer(vr.Version, license.Version) {
		return
	}

	binaryURL := vr.Binaries[binaryKey()]
	n.Notify(vr.Version, vr.ReleaseNotes, vr.PublishedAt)
	ScheduleUpdateWith(ctx, isIdle, p, vr.Version, binaryURL, func() {
		RunUpdate(ctx, vr.Version, binaryURL, dir, p, n)
	})
}

// ScheduleUpdateWith decides whether to run the update now or defer it.
// When idle, it waits 30 seconds (cancellable via ctx) then calls runFn.
// When not idle, it saves the pending update for the next startup.
// Pass nil for runFn to test only the save-pending path.
func ScheduleUpdateWith(ctx context.Context, isIdle func() bool, p *Pending, version, binaryURL string, runFn func()) {
	if isIdle() {
		if runFn != nil {
			go func() {
				select {
				case <-time.After(30 * time.Second):
					runFn()
				case <-ctx.Done():
				}
			}()
		}
	} else {
		p.Save(version, binaryURL)
	}
}
