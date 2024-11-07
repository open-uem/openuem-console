package handlers

import (
	"github.com/doncicuto/openuem-console/internal/views/printers_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) NetworkPrinters(c echo.Context) error {
	return RenderView(c, printers_views.PrintersIndex("| Network Printers", printers_views.Printers()))
}
