package core

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	squareclient "github.com/square/square-go-sdk/client"
	"github.com/team-swsd/circlehq/internal/catalog"
	"github.com/team-swsd/circlehq/internal/model"
	"github.com/team-swsd/circlehq/internal/renderer"
)

var (
	jst *time.Location
)

type Core struct {
	logger           *slog.Logger
	catalog          *catalog.Catalog
	discordClient    DiscordClientInterface
	squareClient     *squareclient.Client
	templateRenderer renderer.TemplateRenderer
}

var _ CoreInterface = (*Core)(nil)

type CoreInterface interface {
	// HandleInventoryUpdateWebhook はSquareの在庫更新Webhookを処理します。
	// サーバーレイヤーはHTTPリクエストのボディをそのまま渡すだけで良いように、ペイロードは[]byteで受け取るのがシンプルです。
	HandleInventoryUpdateWebhook(ctx context.Context, payload []byte) error

	// GetDashboardData はダッシュボード表示用の現在の在庫状況サマリーを取得します。
	// 戻り値は、JSONにシリアライズしやすいように、別途定義した構造体（例: model.Dashboard）が望ましいです。
	GetDashboardData(ctx context.Context) (*model.Dashboard, error)

	RenderDashboardPage(ctx context.Context, w http.ResponseWriter) error
}

type DiscordClientInterface interface {
	Post(ctx context.Context, content string) error
}

func NewCore(logger *slog.Logger, catalog *catalog.Catalog, discordClient DiscordClientInterface, squareClient *squareclient.Client) *Core {
	renderer := renderer.NewHTMLTemplateRenderer()
	return &Core{
		logger:           logger,
		catalog:          catalog,
		discordClient:    discordClient,
		squareClient:     squareClient,
		templateRenderer: renderer,
	}
}

func init() {
	jst, _ = time.LoadLocation("Asia/Tokyo")
}
