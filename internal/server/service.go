package server

import (
	"embed"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/team-swsd/circlehq/internal/core"
)

// service.go

// SSEClient は、各SSEクライアントへのデータ送信用チャネルを表します。
type SSEClient chan []byte

// CircleHQService is a CircleHQ application service
type CircleHQService struct {
	logger       *slog.Logger
	core         core.CoreInterface
	signatureKey string

	// muはsseClientsマップを複数のゴルーチンから安全にアクセスするために使用します。
	mu sync.Mutex
	// sseClientsは、接続している全てのSSEクライアントのチャネルを保持します。
	sseClients map[SSEClient]bool
}

var _ ServerInterface = (*CircleHQService)(nil)

var (
	jst *time.Location
	mu  sync.Mutex // Square Webhook用
)

// NewCircleHQService creates new CircleHQ service
func NewCircleHQService(logger *slog.Logger, core core.CoreInterface, signatureKey string) *CircleHQService {
	jst, _ = time.LoadLocation("Asia/Tokyo")
	return &CircleHQService{
		logger:       logger,
		core:         core,
		signatureKey: signatureKey,
		sseClients:   make(map[SSEClient]bool),
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
func (chqs *CircleHQService) IndexPage(w http.ResponseWriter, r *http.Request) {}

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
