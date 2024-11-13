package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_nats"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) UpdateAgents(c echo.Context) error {
	var err error
	var agents []*ent.Agent

	// Get latest version
	// TODO be able to select channel
	version, err := GetLatestVersion("stable")
	if err != nil {
		log.Println("[INFO]: could not get latest version information")
	}

	if c.Request().Method == "POST" {
		// TODO agent selection, right now is hardcoded for my test
		agents := c.FormValue("agents")
		if agents == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "admin.update.agents.agents_cant_be_empty"), false))
		}

		if version == nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "admin.update.agents.get_version_error"), false))
		}

		updateRequest := openuem_nats.OpenUEMUpdateRequest{}
		updateRequest.DownloadFrom = version.FileURL
		updateRequest.DownloadHash = version.Checksum

		if c.FormValue("update-agent-date") == "" {
			updateRequest.UpdateNow = true
		} else {
			scheduledTime := c.FormValue("update-agent-date")
			updateRequest.UpdateAt, err = time.ParseInLocation("2006-01-02T15:04", scheduledTime, time.Local)
			if err != nil {
				log.Println("[INFO]: could not parse scheduled time as 24h time")
				updateRequest.UpdateAt, err = time.Parse("2006-01-02T15:04PM", scheduledTime)
				if err != nil {
					log.Println("[INFO]: could not parse scheduled time as AM/PM time")

					// Fallback to update now
					updateRequest.UpdateNow = true
				}
			}
		}

		for _, a := range strings.Split(agents, ",") {
			data, err := json.Marshal(updateRequest)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			if _, err := h.NATSConnection.Request("agentupdate."+a, data, 10*time.Second); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "admin.update.agents.success")))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Get filters values
	f := filters.AgentFilter{}
	f.Hostname = c.FormValue("filterByHostname")

	versions, err := h.Model.GetAgentsVersions()
	if err != nil {
		return err
	}
	filteredVersions := []string{}
	for index := range versions {
		value := c.FormValue(fmt.Sprintf("filterByVersion%d", index))
		if value != "" {
			filteredVersions = append(filteredVersions, value)
		}
	}
	f.Versions = filteredVersions

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

	nSelectedItems := c.FormValue("filterBySelectedItems")
	f.SelectedItems, err = strconv.Atoi(nSelectedItems)
	if err != nil {
		f.SelectedItems = 0
	}

	tmpAllAgents := []string{}
	allAgents, err := h.Model.GetAllAgents(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for _, a := range allAgents {
		tmpAllAgents = append(tmpAllAgents, "\""+a.ID+"\"")
	}
	f.SelectedAllAgents = "[" + strings.Join(tmpAllAgents, ",") + "]"

	agents, err = h.Model.GetAgentsByPage(p, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	appliedTags, err := h.Model.GetAppliedTags()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllAgents(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	settings, err := h.Model.GetGeneralSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	higherVersion, err := h.Model.GetHigherAgentVersionInstalled()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.UpdateAgentsIndex(" | Update Agents", admin_views.UpdateAgents(c, p, f, agents, settings, version, higherVersion, versions, availableOSes, appliedTags)))
}

func (h *Handler) UpdateAgentsConfirm(c echo.Context) error {
	version := c.FormValue("version")
	return RenderConfirm(c, partials.ConfirmUpdateAgents(version))
}

func (h *Handler) RollbackAgents(c echo.Context) error {
	var err error

	// TODO agent selection, right now is hardcoded for my test
	agents := c.FormValue("agents")
	if agents == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "admin.update.agents.agents_cant_be_empty"), false))
	}

	rollbackRequest := openuem_nats.OpenUEMRollbackRequest{}

	if c.FormValue("rollback-agent-date") == "" {
		rollbackRequest.RollbackNow = true
	} else {
		scheduledTime := c.FormValue("update-agent-date")
		rollbackRequest.RollbackAt, err = time.ParseInLocation("2006-01-02T15:04", scheduledTime, time.Local)
		if err != nil {
			log.Println("[INFO]: could not parse scheduled time as 24h time")
			rollbackRequest.RollbackAt, err = time.Parse("2006-01-02T15:04PM", scheduledTime)
			if err != nil {
				log.Println("[INFO]: could not parse scheduled time as AM/PM time")

				// Fallback to update now
				rollbackRequest.RollbackNow = true
			}
		}
	}

	for _, a := range strings.Split(agents, ",") {
		data, err := json.Marshal(rollbackRequest)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		if _, err := h.NATSConnection.Request("agentrollback."+a, data, 10*time.Second); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "admin.update.agents.rollback_success")))

}

func (h *Handler) RollbackAgentsConfirm(c echo.Context) error {
	return RenderConfirm(c, partials.ConfirmRollbackAgents())
}

func GetLatestVersion(channel string) (*admin_views.Version, error) {
	// TODO specify the channel
	url := "http://localhost:8888/" + channel

	client := http.Client{
		Timeout: time.Second * 8,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "openuem-console")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	version := admin_views.Version{}
	if err := json.Unmarshal(body, &version); err != nil {
		return nil, err
	}

	return &version, nil
}
