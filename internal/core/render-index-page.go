package core

import (
	"context"
	"net/http"
)

// RenderIndexPage renders the index page using the template renderer.
func (c *Core) RenderIndexPage(ctx context.Context, w http.ResponseWriter, spreadSheetURL string) error {
	c.logger.InfoContext(ctx, "RenderIndexPage")

	// レンダラ
	if err := c.templateRenderer.RenderIndexPage(ctx, w, spreadSheetURL); err != nil {
		c.logger.ErrorContext(ctx, "Failed to render index page", "error", err)
		return err
	}
	return nil
}
