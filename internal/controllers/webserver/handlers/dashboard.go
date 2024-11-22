package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/doncicuto/openuem-console/internal/views"
	"github.com/doncicuto/openuem-console/internal/views/charts"
	"github.com/doncicuto/openuem-console/internal/views/dashboard_views"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Dashboard(c echo.Context) error {
	var err error
	data := dashboard_views.DashboardData{}

	// Get latest version
	channel, err := h.Model.GetDefaultUpdateChannel()
	if err != nil {
		log.Println("[ERROR]: could not get updates channel settings")
		channel = "stable"
	}

	release, err := h.GetLatestRelease(channel)
	if err != nil {
		log.Println("[ERROR]: could not get latest version information")
		data.OpenUEMUpdaterAPIStatus = "down"
		data.NUpgradableAgents = 0
	} else {
		data.OpenUEMUpdaterAPIStatus = "up"
		data.NUpgradableAgents, err = h.Model.CountUpgradableAgents(release.Version)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	data.Charts, err = h.generateCharts(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NOutdatedVersions, err = h.Model.CountOutdatedAgents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NPendingUpdates, err = h.Model.CountPendingUpdateAgents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NInactiveAntiviri, err = h.Model.CountDisabledAntivirusAgents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NOutdatedDatabaseAntiviri, err = h.Model.CountOutdatedAntivirusDatabaseAgents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NNoAutoUpdate, err = h.Model.CountNoAutoupdateAgents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NSupportedVNC, err = h.Model.CountVNCSupportedAgents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NVendors, err = h.Model.CountDifferentVendor()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NPrinters, err = h.Model.CountDifferentPrinters()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	appliedTags, err := h.Model.GetAppliedTags()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	data.NAppliedTags = len(appliedTags)

	data.NDisabledAgents, err = h.Model.CountDisabledAgents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NApps, err = h.Model.CountAllApps(filters.ApplicationsFilter{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NDeployments, err = h.Model.CountAllDeployments()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NOpenUEMUsers, err = h.Model.CountAllUsers(filters.UserFilter{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NSessions, err = h.Model.CountAllSessions()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NUsernames, err = h.Model.CountAllOSUsernames()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.RefreshTime, err = h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		data.RefreshTime = 5
	}

	data.NAgentsNotReportedIn24h, err = h.Model.CountAgentsNotReportedLast24h()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		data.RefreshTime = 5
	}

	data.NCertificatesAboutToExpire, err = h.Model.CountCertificatesAboutToexpire()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.CheckNATSComponentStatus(&data)

	l := views.GetTranslatorForDates(c)

	return RenderView(c, dashboard_views.DashboardIndex("| Dashboard", dashboard_views.Dashboard(h.SessionManager, l, data)))
}

func (h *Handler) generateCharts(c echo.Context) (*dashboard_views.DashboardCharts, error) {
	ch := dashboard_views.DashboardCharts{}

	countAllAgents, err := h.Model.CountAllAgents(filters.AgentFilter{})
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

func (h *Handler) CheckNATSComponentStatus(data *dashboard_views.DashboardData) {

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		data.NATSServerStatus = "down"
		data.AgentWorkerStatus = "down"
		data.NotificationWorkerStatus = "down"
		data.CertManagerWorkerStatus = "down"
	} else {
		data.NATSServerStatus = "up"
		data.AgentWorkerStatus = "up"
		data.NotificationWorkerStatus = "up"
		data.CertManagerWorkerStatus = "up"

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.NATSConnection.Request("ping.agentworker", nil, 1*time.Second); err != nil {
				data.AgentWorkerStatus = "down"
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.NATSConnection.Request("ping.notificationworker", nil, 1*time.Second); err != nil {
				data.NotificationWorkerStatus = "down"
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.NATSConnection.Request("ping.certmanagerworker", nil, 1*time.Second); err != nil {
				data.CertManagerWorkerStatus = "down"
			}
		}()

		wg.Wait()
	}
}
