package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/team-swsd/circlehq/internal/catalog"
	"github.com/team-swsd/circlehq/internal/model"
)

// HandleInventoryUpdateWebhook の実装
func (c *Core) HandleInventoryUpdateWebhook(ctx context.Context, bodyBytes []byte) error {
	// 1. payload ([]byte) をSquareのWebhook構造体にUnmarshalする
	var payload model.InventoryCountUpdatedWebhook
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal webhook payload: %w", err)
	}
	if payload.Type != "inventory.count.updated" {
		// ignore
		return nil
	}

	// 2. catalogを更新
	// 1件の異常データでWebhook全体を落とすと通知も止まってしまうため、
	// パース不能・カタログ未登録のバリエーションはログを残してスキップする。
	for _, v := range payload.Data.Object.InventoryCounts {
		// quantity string to int
		quantity, err := strconv.Atoi(v.Quantity)
		if err != nil {
			c.logger.WarnContext(ctx, "failed to parse inventory quantity; skipping",
				"variation_id", v.CatalogObjectID, "quantity", v.Quantity, "error", err)
			continue
		}
		if err := c.catalog.UpdateQuantityByVariationID(v.CatalogObjectID, quantity); err != nil {
			c.logger.WarnContext(ctx, "inventory update for unknown variation; skipping",
				"variation_id", v.CatalogObjectID, "error", err)
			continue
		}
	}

	// 3. Discordクライアントを使って通知する
	content := c.CreatePostContentText(payload.Data.Object.InventoryCounts)
	if err := c.discordClient.Post(ctx, content); err != nil {
		return err
	}

	// 4. SSEブロードキャスターを使って最新のカタログを配信する
	c.broadcastCatalog(ctx, "webhook_received")

	c.logger.InfoContext(ctx, "Inventory Updated", "catalog", c.catalog)
	return nil
}

// broadcastCatalog は現在のカタログ状態をSSEで全ダッシュボードクライアントへ配信します。
// Webhook受信時と手動リコンサイル時で共通に使います。
// 配信に失敗しても呼び出し元の処理は続行できるよう、エラーはログのみで握ります。
func (c *Core) broadcastCatalog(ctx context.Context, eventType string) {
	dashboardData := map[string]interface{}{
		"type":      eventType,
		"timestamp": time.Now().Unix(),
		"payload":   c.catalog,
	}

	jsonData, err := json.Marshal(dashboardData)
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to marshal dashboard data", "error", err)
		return
	}
	c.logger.InfoContext(ctx, "Broadcasting catalog to dashboard clients", "event", eventType)
	c.broadcaster.Broadcast(jsonData)
}

// ReconcileInventory はSquareから在庫数を取り直してキャッシュを補正し、
// 最新状態をSSEでダッシュボードへ配信します。
func (c *Core) ReconcileInventory(ctx context.Context) error {
	if err := c.catalog.RefreshQuantities(ctx, c.squareClient); err != nil {
		return err
	}
	c.broadcastCatalog(ctx, "reconciled")
	return nil
}

// CreatePostContentText は、InventoryCountのスライスとCatalog情報から投稿用の文字列を生成します。
func (c *Core) CreatePostContentText(inventoryCounts []model.InventoryCount) string {
	// Step 1: VariationのIDから、それが属する親Itemの情報を引くためのマップを作成します。
	// これにより、後の処理で高速にItem情報を参照できます。
	variationToItemMap := make(map[string]catalog.CatalogItem)
	for _, item := range c.catalog.Items {
		for _, variation := range item.Variations {
			variationToItemMap[variation.ID] = item
		}
	}

	// Step 2: 更新情報をItem IDごとにグループ化します。
	// mapのキーをItemのID、値をそのItemに属するInventoryCountのスライスにします。
	itemGroup := make(map[string][]model.InventoryCount)
	for _, count := range inventoryCounts {
		variationID := count.CatalogObjectID
		// マップを使って、VariationがどのItemに属するかを調べます。
		if parentItem, ok := variationToItemMap[variationID]; ok {
			// 対応するItem IDのキーに、このcountを追加します。
			itemGroup[parentItem.ID] = append(itemGroup[parentItem.ID], count)
		}
	}

	// Step 3: グループ化された情報をもとに、Itemごとに投稿文字列を作成します。
	var postContents []string
	for _, countsInGroup := range itemGroup {
		var details []string
		for _, count := range countsInGroup {
			// GetVariationNameByVariationIDを使ってVariationの名前を取得します。
			variationName, _ := c.catalog.GetVariationNameByVariationID(count.CatalogObjectID)
			if variationName == "" {
				variationName = "不明なバリエーション" // 念のため、見つからなかった場合のフォールバック
			}
			// "<VariationName>: 残部<Quantity>" の部分を作成します。
			details = append(details, fmt.Sprintf("%s: 残部%s", variationName, count.Quantity))
		}

		// ヘッダー部分「売れました! <Item名>」を作成します。
		itemName := variationToItemMap[countsInGroup[0].CatalogObjectID].Name
		header := fmt.Sprintf("売れました! %s", itemName)

		// ヘッダーと詳細部分を結合して、一つのItemに関するメッセージを完成させます。
		// 例: "売れました! Tシャツ\nSサイズ: 残部99, Mサイズ: 残部148"
		body := strings.Join(details, ", ")
		postContents = append(postContents, fmt.Sprintf("%s\n%s", header, body))
	}

	// Step 4: 全てのItemのメッセージを改行で連結して、最終的な1つの文字列として返します。
	return strings.Join(postContents, "\n")
}
