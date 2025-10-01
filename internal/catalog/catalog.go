package catalog

import (
	"context"
	"fmt"
	"strconv"

	squaregosdk "github.com/square/square-go-sdk"
	squareclient "github.com/square/square-go-sdk/client"
)

type Catalog struct {
	Items []CatalogItem `json:"items"`
}

type CatalogItem struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
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

func NewCatalog(client *squareclient.Client) (*Catalog, error) {
	catalog, err := getCatalogList(client)
	if err != nil {
		return nil, err
	}

	for i := range catalog.Items {
		for j := range catalog.Items[i].Variations {
			quantity, err := catalog.GetInitialQuantity(client, catalog.Items[i].Variations[j].ID)
			if err != nil {
				return nil, err
			}
			catalog.Items[i].Variations[j].Quantity = quantity
		}
	}
	return catalog, nil
}

func getCatalogList(client *squareclient.Client) (*Catalog, error) {
	var catalog Catalog
	req := &squaregosdk.CatalogListRequest{}
	response, err := client.Catalog.List(context.Background(), req)

	if err != nil {
		return nil, err
	}

	for _, item := range response.Results {

		// アーカイブだったら除外
		if *item.GetItem().ItemData.IsArchived {
			continue
		}

		// アイテムを組み立てる
		catalogItem := CatalogItem{
			ID:   item.GetItem().ID,
			Name: *item.GetItem().ItemData.Name,
		}

		// Catalog Object ID
		variations := item.GetItem().ItemData.GetVariations()

		// バリエーションごとに持ってくる
		catalogItem.Variations = make([]Variation, len(variations))
		for i, v := range variations {
			catalogItem.Variations[i] = Variation{
				ID:       v.ItemVariation.ID,
				Name:     *v.ItemVariation.ItemVariationData.GetName(),
				SKU:      *v.ItemVariation.ItemVariationData.GetSku(),
				Price:    *v.ItemVariation.ItemVariationData.PriceMoney.GetAmount(),
				Sellable: *v.ItemVariation.ItemVariationData.GetSellable(),
			}
		}
		catalog.Items = append(catalog.Items, catalogItem)

	}
	return &catalog, nil
}

func (c *Catalog) GetInitialQuantity(client *squareclient.Client, variationID string) (int, error) {
	req := &squaregosdk.GetInventoryRequest{
		CatalogObjectID: variationID,
	}

	response, err := client.Inventory.Get(context.Background(), req)
	if err != nil {
		return 0, err
	}
	if len(response.Results) == 0 {
		return 0, fmt.Errorf("no inventory found for item %s", variationID)
	}
	item := response.Results[0]
	quantityStr := item.GetQuantity()
	quantity, err := strconv.Atoi(*quantityStr)
	if err != nil {
		return 0, err
	}
	return quantity, nil
}

// UpdateQuantityByVariationID は、指定されたvariationIDを持つVariationのQuantityを更新
// レシーバーをポインタ型 (*Catalog) にすることで、メソッド内での変更が呼び出し元のオブジェクトに反映
func (c *Catalog) UpdateQuantityByVariationID(variationID string, quantity int) error {
	// Catalogが持つすべてのItemをループ
	for i := range c.Items {
		// 各Itemが持つすべてのVariationをループ
		for j := range c.Items[i].Variations {
			// VariationのIDが引数で受け取ったvariationIDと一致するかをチェック
			if c.Items[i].Variations[j].ID == variationID {
				// IDが一致した場合、そのVariationのQuantityを新しい値で更新
				c.Items[i].Variations[j].Quantity = quantity
				// 更新が完了したので、これ以上ループを続ける必要なし。
				return nil
			}
		}
	}
	return nil
}

// GetVariationNameByVariationID は、指定されたvariationIDを持つVariationのNameを返します。
// このメソッドは読み取り専用で構造体を変更しないため、値レシーバー (c Catalog) を使用しています。
// 見つかった場合はNameとtrueを、見つからなかった場合は空文字列とfalseを返します。
func (c *Catalog) GetVariationNameByVariationID(variationID string) (string, bool) {
	// Catalogが持つすべてのItemをループします。
	for _, item := range c.Items {
		// 各Itemが持つすべてのVariationをループします。
		for _, variation := range item.Variations {
			// VariationのIDが引数で受け取ったvariationIDと一致するかをチェックします。
			if variation.ID == variationID {
				// IDが一致した場合、そのVariationのNameとtrueを返します。
				return variation.Name, true
			}
		}
	}

	// 最後までループしても見つからなかった場合、空文字列とfalseを返します。
	return "", false
}
