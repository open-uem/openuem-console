package handlers

import (
	"github.com/labstack/echo/v4"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/printers_views"
)

func (h *Handler) NetworkPrinters(c echo.Context) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	
	return RenderView(c, printers_views.PrintersIndex("| Network Printers", printers_views.Printers(h.SessionManager, h.Version, latestServerRelease.Version)))
}
