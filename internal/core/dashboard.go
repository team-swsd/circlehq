package core

import (
	"context"

	"github.com/team-swsd/circlehq/internal/model"
)

// GetDashboardData の実装
func (c *Core) GetDashboardData(ctx context.Context) (*model.Dashboard, error) {
	// 1. Squareクライアントを使って全アイテムの在庫情報を取得する
	// 2. catalogの情報とマージする
	// 3. ダッシュボード用のデータ構造(model.Dashboard)に整形して返す
	// ...
	return &model.Dashboard{}, nil
}
