package handlers

import (
	"net/http"

	"github.com/doncicuto/openuem-console/internal/views/agents_views"
	"github.com/doncicuto/openuem-console/internal/views/charts"
	"github.com/doncicuto/openuem-console/internal/views/dashboard_views"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Dashboard(c echo.Context) error {
	myCharts, err := h.generateCharts(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return renderView(c, dashboard_views.DashboardIndex("| Dashboard", dashboard_views.Dashboard(*myCharts)))
}

func (h *Handler) generateCharts(c echo.Context) (*dashboard_views.DashboardCharts, error) {
	ch := dashboard_views.DashboardCharts{}

	countAllAgents, err := h.Model.CountAllAgents(agents_views.AgentFilter{})
	if err != nil {
		return nil, err
	}

	topApps, err := h.Model.GetTop10InstalledApps()
	if err != nil {
		return nil, err
	}
	ch.Top10Apps = charts.Top10Apps(topApps)

	agents, err := h.Model.CountAgentsByOS()
	if err != nil {
		return nil, err
	}
	ch.AgentByOs = charts.AgentsByOs(c.Request().Context(), agents, countAllAgents)

	agents, err = h.Model.CountAgentsByOSVersion()
	if err != nil {
		return nil, err
	}

	ch.AgentByOsVersion = charts.AgentsByOsVersion(c.Request().Context(), agents, countAllAgents)

	countAgents, err := h.Model.CountAgentsReportedLast24h()
	if err != nil {
		return nil, err
	}

	ch.AgentByLastReport = charts.AgentsByLastReportDate(c.Request().Context(), countAgents, countAllAgents)

	agents, err = h.Model.CountAgentsByWindowsUpdateStatus()
	if err != nil {
		return nil, err
	}

	ch.AgentBySystemUpdate = charts.AgentsBySystemUpdate(c.Request().Context(), agents, countAllAgents)

	return &ch, nil
}
