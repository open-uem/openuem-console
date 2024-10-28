package handlers

import (
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"

	"github.com/labstack/echo/v4"
)

func (h *Handler) RenderMarkdown(c echo.Context) error {
	// Get markdown

	md := c.FormValue("markdown")

	maybeUnsafeHTML := markdown.ToHTML([]byte(md), nil, nil)
	html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)

	return c.HTML(200, string(html))
}
