package renderer

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/team-swsd/circlehq/internal/model"
)

//go:embed templates/*.html
var tmplFS embed.FS

// templateRenderer is an interface for rendering templates.
type TemplateRenderer interface {
	RenderDashboardPage(ctx context.Context, w http.ResponseWriter, content model.Dashboard) error
}

// Example implementation of TemplateRenderer
type HTMLTemplateRenderer struct{}

func (r *HTMLTemplateRenderer) RenderDashboardPage(ctx context.Context, w http.ResponseWriter, content model.Dashboard) error {
	tmpl := template.New("dashboard.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"now": func() string { return time.Now().Format(time.RFC3339) },
	})
	tmpl, err := tmpl.ParseFS(tmplFS, "templates/dashboard.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	if err := tmpl.Execute(w, content); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return err
}

func NewHTMLTemplateRenderer() TemplateRenderer {
	return &HTMLTemplateRenderer{}
}
