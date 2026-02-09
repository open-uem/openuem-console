package handlers

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/charts"
	"github.com/open-uem/openuem-console/internal/views/dashboard_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
)

func (h *Handler) Dashboard(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	data := dashboard_views.DashboardData{}

	// Get latest version
	channel, err := h.Model.GetDefaultUpdateChannel()
	if err != nil {
		log.Println("[ERROR]: could not get updates channel settings")
		channel = "stable"
	}

	r, err := h.Model.GetLatestAgentRelease(channel)
	if err != nil {
		log.Println("[ERROR]: could not get latest version information")
		data.OpenUEMUpdaterAPIStatus = "down"
		data.NUpgradableAgents = 0
	} else {
		data.OpenUEMUpdaterAPIStatus = "up"
		data.NUpgradableAgents, err = h.Model.CountUpgradableAgents(r.Version)
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

	data.NPendingUpdates, err = h.Model.CountPendingUpdateAgents(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NInactiveAntiviri, err = h.Model.CountDisabledAntivirusAgents(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NOutdatedDatabaseAntiviri, err = h.Model.CountOutdatedAntivirusDatabaseAgents(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NNoAutoUpdate, err = h.Model.CountNoAutoupdateAgents(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NSupportedVNC, err = h.Model.CountVNCSupportedAgents(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NVendors, err = h.Model.CountDifferentVendor(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NPrinters, err = h.Model.CountDifferentPrinters(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	appliedTags, err := h.Model.GetAppliedTags(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	data.NAppliedTags = len(appliedTags)

	data.NDisabledAgents, err = h.Model.CountDisabledAgents(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NWaitingForAdmission, err = h.Model.CountWaitingForAdmissionAgents(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NApps, err = h.Model.CountAllApps(filters.ApplicationsFilter{}, commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NDeployments, err = h.Model.CountAllDeployments(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if tenantID == -1 {
		// Global admin view - use main tenant
		mainTenant, err := h.Model.GetMainTenant()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		tenantID = mainTenant.ID
	}

	data.NOpenUEMUsers, err = h.Model.CountAllUsers(filters.UserFilter{}, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NSessions, err = h.Model.CountAllSessions()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.NUsernames, err = h.Model.CountAllOSUsernames(commonInfo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.RefreshTime, err = h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		data.RefreshTime = 5
	}

	data.NAgentsNotReportedIn24h, err = h.Model.CountAgentsNotReportedLast24h(commonInfo)
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		data.RefreshTime = 5
	}

	data.NCertificatesAboutToExpire, err = h.Model.CountCertificatesAboutToexpire()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.CheckNATSComponentStatus(&data)

	return RenderView(c, dashboard_views.DashboardIndex("| Dashboard", dashboard_views.Dashboard(c, data, commonInfo), commonInfo))
}

func (h *Handler) generateCharts(c echo.Context) (*dashboard_views.DashboardCharts, error) {
	ch := dashboard_views.DashboardCharts{}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return nil, err
	}

	countAllAgents, err := h.Model.CountAllAgents(filters.AgentFilter{}, true, commonInfo)
	if err != nil {
		return nil, err
	}

	agents, err := h.Model.CountAgentsByOS(commonInfo)
	if err != nil {
		return nil, err
	}
	ch.AgentByOs = charts.AgentsByOs(c.Request().Context(), agents, countAllAgents)

	agents, err = h.Model.CountAgentsByOSVersion(commonInfo)
	if err != nil {
		return nil, err
	}

	ch.AgentByOsVersion = charts.AgentsByOsVersion(c.Request().Context(), agents, countAllAgents)

	countAgents, err := h.Model.CountAgentsReportedLast24h(commonInfo)
	if err != nil {
		return nil, err
	}

	ch.AgentByLastReport = charts.AgentsByLastReportDate(c.Request().Context(), countAgents, countAllAgents)

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
