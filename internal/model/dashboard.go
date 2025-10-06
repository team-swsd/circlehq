package model

type Dashboard struct {
	Items     []DashboardItem `json:"items"`
	UpdatedAt string          `json:"updated_at"`
}
type DashboardItem struct {
	Name       string      `json:"name"`
	ID         string      `json:"id"`
	Variations []Variation `json:"variations"`
}

type Variation struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SKU      string `json:"sku"`
	Price    int64  `json:"price"`
	Sellable bool   `json:"sellable"`
	Quantity int    `json:"quantity"`
}
