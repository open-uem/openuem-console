package handlers

import (
	"fmt"
	"log"

	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem-console/internal/views/security_views"
	"github.com/doncicuto/openuem_nats"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListAntivirusStatus(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Get filters values
	f := filters.AntivirusFilter{}
	f.Hostname = c.FormValue("filterByHostname")

	availableOSes, err := h.Model.GetAgentsUsedOSes()
	if err != nil {
		return err
	}
	filteredAgentOSes := []string{}
	for index := range availableOSes {
		value := c.FormValue(fmt.Sprintf("filterByAgentOS%d", index))
		if value != "" {
			filteredAgentOSes = append(filteredAgentOSes, value)
		}
	}
	f.AgentOSVersions = filteredAgentOSes

	detectedAntiviri, err := h.Model.GetDetectedAntiviri()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	filteredAntiviri := []string{}
	for index := range detectedAntiviri {
		value := c.FormValue(fmt.Sprintf("filterByAntivirusName%d", index))
		if value != "" {
			filteredAntiviri = append(filteredAntiviri, value)
		}
	}
	f.AntivirusNameOptions = filteredAntiviri

	filteredEnableStatus := []string{}
	for index := range []string{"Enabled", "Disabled"} {
		value := c.FormValue(fmt.Sprintf("filterByAntivirusEnabled%d", index))
		if value != "" {
			filteredEnableStatus = append(filteredEnableStatus, value)
		}
	}
	f.AntivirusEnabledOptions = filteredEnableStatus

	filteredUpdateStatus := []string{}
	for index := range []string{"UpdatedYes", "UpdatedNo"} {
		value := c.FormValue(fmt.Sprintf("filterByAntivirusUpdated%d", index))
		if value != "" {
			filteredUpdateStatus = append(filteredUpdateStatus, value)
		}
	}
	f.AntivirusUpdatedOptions = filteredUpdateStatus

	antiviri, err := h.Model.GetAntiviriByPage(p, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllAntiviri(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	return RenderView(c, security_views.SecurityIndex("| Security", security_views.Antivirus(c, p, f, antiviri, detectedAntiviri, availableOSes, refreshTime)))
}

func (h *Handler) ListSecurityUpdatesStatus(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Get filters values
	f := filters.SystemUpdatesFilter{}
	f.Hostname = c.FormValue("filterByHostname")

	availableOSes, err := h.Model.GetAgentsUsedOSes()
	if err != nil {
		return err
	}
	filteredAgentOSes := []string{}
	for index := range availableOSes {
		value := c.FormValue(fmt.Sprintf("filterByAgentOS%d", index))
		if value != "" {
			filteredAgentOSes = append(filteredAgentOSes, value)
		}
	}
	f.AgentOSVersions = filteredAgentOSes

	lastSearchFrom := c.FormValue("filterLastSearchDateFrom")
	if lastSearchFrom != "" {
		f.LastSearchFrom = lastSearchFrom
	}
	lastSearchTo := c.FormValue("filterLastSearchDateTo")
	if lastSearchTo != "" {
		f.LastSearchTo = lastSearchTo
	}

	lastInstallFrom := c.FormValue("filterLastInstallDateFrom")
	if lastInstallFrom != "" {
		f.LastInstallFrom = lastInstallFrom
	}
	lastInstallTo := c.FormValue("filterLastInstallDateTo")
	if lastInstallTo != "" {
		f.LastInstallTo = lastInstallTo
	}

	filteredPendingUpdates := []string{}
	for index := range []string{"Yes", "No"} {
		value := c.FormValue(fmt.Sprintf("filterByPendingUpdate%d", index))
		if value != "" {
			filteredPendingUpdates = append(filteredPendingUpdates, value)
		}
	}
	f.PendingUpdateOptions = filteredPendingUpdates

	availableUpdateStatus := openuem_nats.SystemUpdatePossibleStatus()
	filteredUpdateStatus := []string{}
	for index := range availableUpdateStatus {
		value := c.FormValue(fmt.Sprintf("filterByUpdateStatus%d", index))
		if value != "" {
			filteredUpdateStatus = append(filteredUpdateStatus, value)
		}
	}
	f.UpdateStatus = filteredUpdateStatus

	systemUpdates, err := h.Model.GetSystemUpdatesByPage(p, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllSystemUpdates(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	return RenderView(c, security_views.SecurityIndex("| Security", security_views.SecurityUpdates(c, p, f, systemUpdates, availableOSes, availableUpdateStatus, refreshTime)))
}

func (h *Handler) ListLatestUpdates(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
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
