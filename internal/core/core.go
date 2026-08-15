package core

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	squareclient "github.com/square/square-go-sdk/client"
	"github.com/team-swsd/circlehq/internal/broadcast"
	"github.com/team-swsd/circlehq/internal/catalog"
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
	broadcaster      *broadcast.Broadcaster
}

var _ CoreInterface = (*Core)(nil)

type CoreInterface interface {
	// HandleInventoryUpdateWebhook はSquareの在庫更新Webhookを処理します。
	// サーバーレイヤーはHTTPリクエストのボディをそのまま渡すだけで良いように、ペイロードは[]byteで受け取るのがシンプルです。
	HandleInventoryUpdateWebhook(ctx context.Context, payload []byte) error

	// ReconcileInventory はSquareから在庫数を取り直してキャッシュを補正し、
	// 結果をSSEでダッシュボードへ反映します（手動更新用）。
	ReconcileInventory(ctx context.Context) error

	RenderDashboardPage(ctx context.Context, w http.ResponseWriter) error
	RenderIndexPage(ctx context.Context, w http.ResponseWriter, spreadSheetURL string) error

	Broadcaster() *broadcast.Broadcaster // このメソッドを追加
}

type DiscordClientInterface interface {
	Post(ctx context.Context, content string) error
}

func NewCore(logger *slog.Logger, catalog *catalog.Catalog, discordClient DiscordClientInterface, squareClient *squareclient.Client, broadcaster *broadcast.Broadcaster) *Core {
	renderer := renderer.NewHTMLTemplateRenderer()
	return &Core{
		logger:           logger,
		catalog:          catalog,
		discordClient:    discordClient,
		squareClient:     squareClient,
		templateRenderer: renderer,
		broadcaster:      broadcaster,
	}
}

// Broadcaster は core が保持する broadcaster インスタンスを返します。
// このメソッドを実装
func (c *Core) Broadcaster() *broadcast.Broadcaster {
	return c.broadcaster
}

func init() {
	jst, _ = time.LoadLocation("Asia/Tokyo")
}
