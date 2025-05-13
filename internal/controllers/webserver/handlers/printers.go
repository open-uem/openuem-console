package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/printers_views"
)

func (h *Handler) NetworkPrinters(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	return RenderView(c, printers_views.PrintersIndex("| Network Printers", printers_views.Printers(c, commonInfo), commonInfo))
}
