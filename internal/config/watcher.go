package config

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches a config file and calls onReload when it changes.
type Watcher struct {
	path     string
	onReload func(*Config)
	debounce time.Duration
}

// NewWatcher creates a Watcher for the given config file path.
func NewWatcher(path string, onReload func(*Config)) (*Watcher, error) {
	return &Watcher{
		path:     path,
		onReload: onReload,
		debounce: 300 * time.Millisecond,
	}, nil
}

// Run starts watching the config file. Blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer fw.Close()

	if err := fw.Add(w.path); err != nil {
		return
	}

	var timer *time.Timer
	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return
		case event, ok := <-fw.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(w.debounce, func() {
					cfg, err := Load(w.path)
					if err != nil {
						return
					}
					w.onReload(cfg)
				})
			}
		case _, ok := <-fw.Errors:
			if !ok {
				return
			}
		}
	}
}
