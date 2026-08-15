package server

import (
	"embed"
	"log/slog"
	"net/http"
	"sync"

	"github.com/team-swsd/circlehq/internal/core"
)

// service.go

// SSEClient は、各SSEクライアントへのデータ送信用チャネルを表します。
type SSEClient chan []byte

// CircleHQService is a CircleHQ application service
type CircleHQService struct {
	logger         *slog.Logger
	core           core.CoreInterface
	signatureKey   string
	spreadSheetURL string
}

var _ ServerInterface = (*CircleHQService)(nil)

var (
	mu sync.Mutex // Square Webhook用
)

// NewCircleHQService creates new CircleHQ service
func NewCircleHQService(logger *slog.Logger, core core.CoreInterface, signatureKey string, spreadSheetURL string) *CircleHQService {
	return &CircleHQService{
		logger:         logger,
		core:           core,
		signatureKey:   signatureKey,
		spreadSheetURL: spreadSheetURL,
	}
}

// Get Health Check
// (GET /health)
func (chqs *CircleHQService) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chqs.logger.InfoContext(ctx, "HealthCheck")
	writeHealthResponse(w)
}

// index
// (GET /)
func (chqs *CircleHQService) IndexPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chqs.logger.InfoContext(ctx, "IndexPage")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := chqs.core.RenderIndexPage(ctx, w, chqs.spreadSheetURL); err != nil {
		chqs.logger.ErrorContext(ctx, "Failed to render index page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Successfully rendered the page
	chqs.logger.InfoContext(ctx, "Index page rendered successfully")
}

// dashboard
// (GET /dashboard)
func (chqs *CircleHQService) DashboardPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chqs.logger.InfoContext(ctx, "DashboardPage")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := chqs.core.RenderDashboardPage(ctx, w); err != nil {
		chqs.logger.ErrorContext(ctx, "Failed to render dashboard page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Successfully rendered the page
	chqs.logger.InfoContext(ctx, "Dashboard page rendered successfully")
}

// Manually reconcile inventory from Square
// (POST /api/reconcile)
func (chqs *CircleHQService) ReconcileInventory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chqs.logger.InfoContext(ctx, "ReconcileInventory")
	if err := chqs.core.ReconcileInventory(ctx); err != nil {
		chqs.logger.ErrorContext(ctx, "Failed to reconcile inventory", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	chqs.logger.InfoContext(ctx, "Inventory reconciled successfully")
}

// google spreadsheet
// (GET /reservation)
func (chqs *CircleHQService) SpreadsheetPage(w http.ResponseWriter, r *http.Request) {}

// static files
//
//go:embed static/*
var embeddedStaticFS embed.FS

// (GET /static/*)
func (chqs *CircleHQService) StaticFiles(w http.ResponseWriter, r *http.Request) {
	httpFS := http.FS(embeddedStaticFS)
	fs := http.FileServer(httpFS)
	// handler := http.StripPrefix("/static/", fs)
	handler := fs
	handler.ServeHTTP(w, r)
}
