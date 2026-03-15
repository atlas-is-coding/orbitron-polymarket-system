package updater

import (
	"context"
	"fmt"

	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/rs/zerolog/log"
)

// Notifier sends update-available events to all channels.
type Notifier struct {
	bus      *tui.EventBus
	telegram notify.Notifier
}

// NewNotifier creates a Notifier. Pass nil for channels you don't need.
func NewNotifier(bus *tui.EventBus, telegram notify.Notifier) *Notifier {
	return &Notifier{bus: bus, telegram: telegram}
}

// Notify fires an update-available event to TUI, Telegram, and log.
func (n *Notifier) Notify(version, releaseNotes, publishedAt string) {
	log.Info().
		Str("latest", version).
		Str("published_at", publishedAt).
		Str("notes", releaseNotes).
		Msg("update available")

	if n.bus != nil {
		n.bus.Send(tui.UpdateAvailableMsg{
			Version:      version,
			ReleaseNotes: releaseNotes,
			PublishedAt:  publishedAt,
		})
	}

	if n.telegram != nil {
		msg := fmt.Sprintf(
			"Orbitron update available: v%s\n%s\nPublished: %s",
			version, releaseNotes, publishedAt,
		)
		if err := n.telegram.Send(context.Background(), msg); err != nil {
			log.Warn().Err(err).Msg("updater: telegram notification failed")
		}
	}
}

// NotifyError sends an error notification via Telegram and log.
func (n *Notifier) NotifyError(msg string) {
	log.Error().Msg("updater: " + msg)
	if n.telegram != nil {
		if err := n.telegram.Send(context.Background(), "Orbitron updater error: "+msg); err != nil {
			log.Warn().Err(err).Msg("updater: telegram error notification failed")
		}
	}
}
