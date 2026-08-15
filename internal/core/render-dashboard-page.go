package core

import (
	"context"
	"net/http"

	"github.com/team-swsd/circlehq/internal/model"
)

// RenderDashboardPage renders the dashboard page using the template renderer.
// 在庫数は起動時取得＋Webhook差分で更新されるオンメモリキャッシュをそのまま使う。
// ページを開くたびにSquare APIを叩くと負荷・レイテンシ・レート制限の面で不利なため、
// 描画はキャッシュのスナップショットのみで行い、ズレの補正は手動リコンサイル(ReconcileInventory)に任せる。
func (c *Core) RenderDashboardPage(ctx context.Context, w http.ResponseWriter) error {
	c.logger.InfoContext(ctx, "RenderDashboardPage")

	// カタログはWebhookで更新され得るので、ロック下のスナップショットを読む
	updatedAt, items := c.catalog.Snapshot()
	dashboardData := model.Dashboard{
		UpdatedAt: updatedAt,
	}
	for _, item := range items {
		itemData := model.DashboardItem{
			Name: item.Name,
			ID:   item.ID,
		}
		for _, variation := range item.Variations {
			itemData.Variations = append(itemData.Variations, model.Variation{
				ID:       variation.ID,
				Name:     variation.Name,
				SKU:      variation.SKU,
				Price:    variation.Price,
				Sellable: variation.Sellable,
				Quantity: variation.Quantity,
			})
		}
		dashboardData.Items = append(dashboardData.Items, itemData)
	}

	// レンダラ
	if err := c.templateRenderer.RenderDashboardPage(ctx, w, dashboardData); err != nil {
		c.logger.ErrorContext(ctx, "Failed to render dashboard page", "error", err)
		return err
	}
	return nil
}
