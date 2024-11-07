package handlers

import (
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem-console/internal/views/security_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListAntivirusStatus(c echo.Context) error {
	agents, err := h.Model.GetAntiviriInfo()
	if err != nil {
		return RenderView(c, security_views.SecurityIndex("| Security", partials.Error(err.Error(), "Security", "/security")))
	}
	return RenderView(c, security_views.SecurityIndex("| Security", security_views.Antivirus(agents)))
}

func (h *Handler) ListSecurityUpdatesStatus(c echo.Context) error {
	agents, err := h.Model.GetSystemUpdatesInfo()
	if err != nil {
		return RenderView(c, security_views.SecurityIndex("| Security", partials.Error(err.Error(), "Security", "/security")))
	}

	return RenderView(c, security_views.SecurityIndex("| Security", security_views.SecurityUpdates(agents)))
}

func (h *Handler) ListLatestUpdates(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage("agent id cannot be empty", false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not find agent info", false))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	p.NItems, err = h.Model.CountLatestUpdates(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	updates, err := h.Model.GetLatestUpdates(agentId, p)
	if err != nil {
		return RenderView(c, security_views.SecurityIndex("| Security", partials.Error(err.Error(), "Security", "/security")))
	}

	if c.Request().Method == "POST" {
		return RenderView(c, security_views.LatestUpdates(c, p, agent, updates))
	}

	return RenderView(c, security_views.SecurityIndex("| Security", security_views.LatestUpdates(c, p, agent, updates)))
}
