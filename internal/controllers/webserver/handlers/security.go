package handlers

import (
	"fmt"
	"log"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/security_views"
)

func (h *Handler) ListAntivirusStatus(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	// Get filters values
	f, availableOSes, detectedAntiviri, err := h.GetAntiviriFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	antiviri, err := h.Model.GetAntiviriByPage(p, *f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllAntiviri(*f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	l := views.GetTranslatorForDates(c)

	return RenderView(c, security_views.SecurityIndex("| Security", security_views.Antivirus(c, p, *f, h.SessionManager, l, antiviri, detectedAntiviri, availableOSes, refreshTime)))
}

func (h *Handler) ListSecurityUpdatesStatus(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	// Get filters values
	f, availableOSes, availableUpdateStatus, err := h.GetSystemUpdatesFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	systemUpdates, err := h.Model.GetSystemUpdatesByPage(p, *f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllSystemUpdates(*f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	l := views.GetTranslatorForDates(c)

	return RenderView(c, security_views.SecurityIndex("| Security", security_views.SecurityUpdates(c, p, *f, h.SessionManager, l, systemUpdates, availableOSes, availableUpdateStatus, refreshTime)))
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
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

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
		return RenderView(c, security_views.SecurityIndex("| Security", partials.Error(err.Error(), "Security", "/security", h.SessionManager)))
	}

	l := views.GetTranslatorForDates(c)

	if c.Request().Method == "POST" {
		return RenderView(c, security_views.LatestUpdates(c, p, h.SessionManager, l, agent, updates))
	}

	return RenderView(c, security_views.SecurityIndex("| Security", security_views.LatestUpdates(c, p, h.SessionManager, l, agent, updates)))
}

func (h *Handler) GetAntiviriFilters(c echo.Context) (*filters.AntivirusFilter, []string, []string, error) {
	// Get filters values
	f := filters.AntivirusFilter{}
	f.Hostname = c.FormValue("filterByHostname")

	availableOSes, err := h.Model.GetAgentsUsedOSes()
	if err != nil {
		return nil, nil, nil, err
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
		return nil, nil, nil, err
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

	return &f, availableOSes, detectedAntiviri, nil
}

func (h *Handler) GetSystemUpdatesFilters(c echo.Context) (*filters.SystemUpdatesFilter, []string, []string, error) {
	// Get filters values
	f := filters.SystemUpdatesFilter{}
	f.Hostname = c.FormValue("filterByHostname")

	availableOSes, err := h.Model.GetAgentsUsedOSes()
	if err != nil {
		return nil, nil, nil, err
	}
	filteredAgentOSes := []string{}
	for index := range availableOSes {
		value := c.FormValue(fmt.Sprintf("filterByAgentOS%d", index))
		if value != "" {
			filteredAgentOSes = append(filteredAgentOSes, value)
		}
	}
	f.AgentOSVersions = filteredAgentOSes

	lastSearchFrom := c.FormValue("filterByLastSearchDateFrom")
	if lastSearchFrom != "" {
		f.LastSearchFrom = lastSearchFrom
	}
	lastSearchTo := c.FormValue("filterByLastSearchDateTo")
	if lastSearchTo != "" {
		f.LastSearchTo = lastSearchTo
	}

	lastInstallFrom := c.FormValue("filterByLastInstallDateFrom")
	if lastInstallFrom != "" {
		f.LastInstallFrom = lastInstallFrom
	}
	lastInstallTo := c.FormValue("filterByLastInstallDateTo")
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

	return &f, availableOSes, availableUpdateStatus, err
}
