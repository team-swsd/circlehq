package model

import "time"

type InventoryCountUpdatedWebhook struct {
	MerchantID string                                `json:"merchant_id"`
	Type       string                                `json:"type"`
	EventID    string                                `json:"event_id"`
	CreatedAt  time.Time                             `json:"created_at"`
	Data       InventoryCountUpdatedWebhookDataBlock `json:"data"`
}

type InventoryCountUpdatedWebhookDataBlock struct {
	Type   string                                        `json:"type"`
	ID     string                                        `json:"id"`
	Object InventoryCountUpdatedWebhookDataObjectWrapper `json:"object"`
}

type InventoryCountUpdatedWebhookDataObjectWrapper struct {
	InventoryCounts []InventoryCount `json:"inventory_counts"`
}

type InventoryCount struct {
	CalculatedAt      time.Time `json:"calculated_at"`
	CatalogObjectID   string    `json:"catalog_object_id"`
	CatalogObjectType string    `json:"catalog_object_type"`
	LocationID        string    `json:"location_id"`
	Quantity          string    `json:"quantity"`
	State             string    `json:"state"`
}
