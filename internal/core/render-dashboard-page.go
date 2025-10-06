package core

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	squaregosdk "github.com/square/square-go-sdk"
	"github.com/team-swsd/circlehq/internal/model"
)

// RenderDashboardPage renders the dashboard page using the template renderer.
func (c *Core) RenderDashboardPage(ctx context.Context, w http.ResponseWriter) error {
	c.logger.InfoContext(ctx, "RenderDashboardPage")

	dashboardData := model.Dashboard{
		UpdatedAt: c.catalog.UpdatedAt,
	}
	for _, item := range c.catalog.Items {
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
			// 在庫情報を取得
			quantity, err := c.getQuantity(variation.ID)
			if err != nil {
				c.logger.ErrorContext(ctx, "Failed to get quantity for variation", "variation_id", variation.ID, "error", err)
				continue // エラーがあっても処理を続ける
			}
			itemData.Variations[len(itemData.Variations)-1].Quantity = quantity
		}
		dashboardData.Items = append(dashboardData.Items, itemData)
	}
	fmt.Println(dashboardData)

	// レンダラ
	if err := c.templateRenderer.RenderDashboardPage(ctx, w, dashboardData); err != nil {
		c.logger.ErrorContext(ctx, "Failed to render dashboard page", "error", err)
		return err
	}
	return nil
}

func (c *Core) getQuantity(itemID string) (int, error) {
	// カタログからアイテムを取得
	req := &squaregosdk.GetInventoryRequest{
		CatalogObjectID: itemID,
	}

	response, err := c.squareClient.Inventory.Get(context.Background(), req)

	if err != nil {
		return 0, fmt.Errorf("failed to get inventory for item %s: %w", itemID, err)
	}

	if len(response.Results) == 0 {
		return 0, fmt.Errorf("no inventory found for item %s", itemID)
	}

	item := response.Results[0]
	quantityStr := item.GetQuantity()
	quantity, err := strconv.Atoi(*quantityStr)
	if err != nil {
		fmt.Println("Error parsing quantity:", err)
		return 0, fmt.Errorf("invalid quantity for item %s: %w", itemID, err)
	}

	if *item.GetCatalogObjectID() != itemID {
		return 0, fmt.Errorf("item ID mismatch: expected %s, got %s", itemID, *item.GetCatalogObjectID())
	}

	return quantity, nil
}
