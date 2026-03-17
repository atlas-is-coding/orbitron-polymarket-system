package webui

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/rs/zerolog"
)

//go:embed web/dist
var staticFiles embed.FS

// New creates a Server. canceler, wallets, and mkts may be nil.
func New(
	cfg *config.Config,
	cfgPath string,
	bus *tui.EventBus,
	nx *tui.Nexus,
	canceler OrderCanceler,
	wallets WalletMutator,
	mkts MarketsProvider,
	placer OrderPlacer,
	trading TradingProvider,
	log *zerolog.Logger,
) *Server {
	s := &Server{
		cfg:      cfg,
		cfgPath:  cfgPath,
		password: cfg.WebUI.JWTSecret,
		bus:      bus,
		nx:       nx,
		canceler: canceler,
		wallets:  wallets,
		mkts:     mkts,
		placer:   placer,
		trading:  trading,
		state:    newWebState(),
		hub:      newHub(),
	}
	s.state.SetConfig(cfg)
	return s
}
// recoverMiddleware catches handler panics and returns a JSON 500 error instead
// of resetting the TCP connection (which browsers report as "Network Error").
func (s *Server) recoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				writeError(w, http.StatusInternalServerError, fmt.Sprintf("internal error: %v", err))
			}
		}()
		next(w, r)
	}
}

// Run starts the HTTP server and EventBus consumer. Blocks until ctx is done.
func (s *Server) Run(ctx context.Context) error {
	// Subscribe to EventBus
	tap := s.bus.Tap()
	go s.hub.consume(ctx, tap, s.nx, s.state)

	mux := http.NewServeMux()

	// Static SPA files with HTML5 history-mode fallback.
	// For any path that is not a real static file, serve index.html so that
	// Vue Router can handle the route on the client side (page refresh on /markets etc.).
	distFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		return fmt.Errorf("webui: embed fs: %w", err)
	}
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}
		f, openErr := distFS.Open(name)
		if openErr == nil {
			stat, statErr := f.Stat()
			f.Close()
			if statErr == nil && !stat.IsDir() {
				// Real static asset — delegate to the standard file server (ETags, range, etc.)
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// SPA route — serve index.html and let Vue Router take over.
		idxBytes, _ := staticFiles.ReadFile("web/dist/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(idxBytes) //nolint:errcheck
	})

	// Auth (no JWT required)
	mux.HandleFunc("/api/v1/login", s.handleLogin)

	// Public: health (no JWT required)
	mux.HandleFunc("/api/v1/health", s.handleGetHealth)

	// Protected: overview, orders, positions, logs
	mux.HandleFunc("/api/v1/overview", s.jwtMiddleware(s.handleOverview))
	mux.HandleFunc("/api/v1/positions", s.jwtMiddleware(s.handlePositions))
	mux.HandleFunc("/api/v1/strategies", s.jwtMiddleware(s.handleStrategies))
	mux.HandleFunc("/api/v1/strategies/", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
	        path := r.URL.Path
	        switch {
	        case strings.HasSuffix(path, "/start"):
	                s.handleStartStrategy(w, r)
	        case strings.HasSuffix(path, "/stop"):
	                s.handleStopStrategy(w, r)
	        default:
	                writeError(w, http.StatusNotFound, "not found")
	        }
	}))
	mux.HandleFunc("/api/v1/logs", s.jwtMiddleware(s.handleLogs))
	// Initial subsystem status for Web UI itself
	s.state.SetSubsystem("Web UI", true)
	// Orders: GET list / POST place / DELETE all — and DELETE single by path suffix
	mux.HandleFunc("/api/v1/orders", s.jwtMiddleware(s.recoverMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleOrders(w, r)
		case http.MethodPost:
			s.handlePlaceOrder(w, r)
		case http.MethodDelete:
			s.handleCancelAll(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})))
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
		path := r.URL.Path
		switch {
		case r.Method == http.MethodDelete:
			s.handleRemoveTrader(w, r)
		case r.Method == http.MethodPatch && strings.HasSuffix(path, "/toggle"):
			s.handleToggleTrader(w, r)
		case r.Method == http.MethodPatch && strings.HasSuffix(path, "/edit"):
			s.handleEditTrader(w, r)
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

	// Wallets
	mux.HandleFunc("/api/v1/wallets", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetWallets(w, r)
		case http.MethodPost:
			s.handleAddWallet(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))
	mux.HandleFunc("/api/v1/wallets/", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPatch:
			s.handleUpdateWallet(w, r)
		case r.Method == http.MethodDelete:
			s.handleDeleteWallet(w, r)
		case r.Method == http.MethodPost:
			s.handleToggleWallet(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))

	// Markets — order matters: exact paths before the wildcard subtree
	mux.HandleFunc("/api/v1/markets/tags", s.jwtMiddleware(s.handleMarketsTags))
	mux.HandleFunc("/api/v1/markets/trending", s.jwtMiddleware(s.handleMarketsTrending))
	mux.HandleFunc("/api/v1/markets/stats", s.jwtMiddleware(s.handleMarketsStats))
	mux.HandleFunc("/api/v1/markets/", s.jwtMiddleware(s.handleMarketDetail))
	mux.HandleFunc("/api/v1/markets", s.jwtMiddleware(s.handleMarketsList))
	mux.HandleFunc("/api/v1/alerts", s.jwtMiddleware(s.handleCreateAlert))

	// WebSocket
	mux.HandleFunc("/ws", s.jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		s.hub.serveWS(w, r, s.nx, s.state, r.RemoteAddr)
	}))
	srv := &http.Server{Addr: s.cfg.WebUI.Listen, Handler: mux}

	go func() {
		<-ctx.Done()
		// Give active requests up to 10 seconds to finish before forcing close.
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			_ = srv.Close()
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("webui: %w", err)
	}
	return nil
}
