package handlers

import (
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem-console/internal/views/security_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListAntivirusStatus(c echo.Context) error {
	agents, err := h.Model.GetAntiviriInfo()
	if err != nil {
		return renderView(c, security_views.SecurityIndex("| Security", partials.Error(err.Error(), "Security", "/security")))
	}
	return renderView(c, security_views.SecurityIndex("| Security", security_views.Antivirus(agents)))
}

func (h *Handler) ListSecurityUpdatesStatus(c echo.Context) error {
	agents, err := h.Model.GetSystemUpdatesInfo()
	if err != nil {
		return renderView(c, security_views.SecurityIndex("| Security", partials.Error(err.Error(), "Security", "/security")))
	}

	return renderView(c, security_views.SecurityIndex("| Security", security_views.SecurityUpdates(agents)))
}
