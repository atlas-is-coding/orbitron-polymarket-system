package webui

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/tui"
	"github.com/rs/zerolog"
)

//go:embed web/dist
var staticFiles embed.FS

// New creates a Server. canceler may be nil if TradesMonitor is disabled.
func New(
	cfg *config.Config,
	cfgPath string,
	bus *tui.EventBus,
	canceler OrderCanceler,
	log *zerolog.Logger,
) *Server {
	s := &Server{
		cfg:      cfg,
		cfgPath:  cfgPath,
		password: cfg.WebUI.JWTSecret,
		bus:      bus,
		canceler: canceler,
		state:    newWebState(),
		hub:      newHub(),
	}
	s.state.SetConfig(cfg)
	return s
}

// Run starts the HTTP server and EventBus consumer. Blocks until ctx is done.
func (s *Server) Run(ctx context.Context) error {
	// Subscribe to EventBus
	tap := s.bus.Tap()
	go s.hub.consume(ctx, tap, s.state)

	mux := http.NewServeMux()

	// Static SPA files
	distFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		return fmt.Errorf("webui: embed fs: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(distFS)))

	// Auth (no JWT required)
	mux.HandleFunc("/api/v1/login", s.handleLogin)

	// Protected: overview, orders, positions, logs
	mux.HandleFunc("/api/v1/overview", s.jwtMiddleware(s.handleOverview))
	mux.HandleFunc("/api/v1/positions", s.jwtMiddleware(s.handlePositions))
	mux.HandleFunc("/api/v1/logs", s.jwtMiddleware(s.handleLogs))

	// Orders: GET list / DELETE all — and DELETE single by path suffix
	mux.HandleFunc("/api/v1/orders", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleOrders(w, r)
		case http.MethodDelete:
			s.handleCancelAll(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))
	mux.HandleFunc("/api/v1/orders/", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			s.handleCancelOrder(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))

	// Copytrading
	mux.HandleFunc("/api/v1/copytrading", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleCopytrading(w, r)
		case http.MethodPost:
			s.handleAddTrader(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))
	mux.HandleFunc("/api/v1/copytrading/traders/", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			s.handleRemoveTrader(w, r)
		case http.MethodPatch:
			s.handleToggleTrader(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))

	// Settings
	mux.HandleFunc("/api/v1/settings", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetSettings(w, r)
		case http.MethodPost:
			s.handlePostSettings(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))

	// WebSocket
	mux.HandleFunc("/ws", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		s.hub.serveWS(w, r, r.RemoteAddr)
	}))

	srv := &http.Server{Addr: s.cfg.WebUI.Listen, Handler: mux}

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("webui: %w", err)
	}
	return nil
}
