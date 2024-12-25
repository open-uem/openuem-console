package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/doncicuto/openuem-console/internal/views"
	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/release"
	"github.com/doncicuto/openuem_nats"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) UpdateAgents(c echo.Context) error {
	var err error

	successMessage := ""
	errorMessage := ""

	// Get latest version
	channel, err := h.Model.GetDefaultUpdateChannel()
	if err != nil {
		log.Println("[ERROR]: could not get updates channel settings")
		channel = "stable"
	}

	r, err := h.Model.GetLatestAgentRelease(channel)
	if err != nil {
		log.Println("[ERROR]: could not get latest version information")
	}

	if c.Request().Method == "POST" {
		agents := c.FormValue("agents")
		if agents == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "admin.update.agents.agents_cant_be_empty"), false))
		}

		sr := c.FormValue("filterBySelectedRelease")
		if sr == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "admin.update.agents.release_cant_be_empty"), false))
		}

		for _, a := range strings.Split(agents, ",") {

			agentInfo, err := h.Model.GetAgentById(a)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			arch := ""
			switch agentInfo.Edges.Computer.ProcessorArch {
			case "x64":
				arch = "amd64"
			}

			releaseToBeApplied, err := h.Model.GetAgentsReleaseByType(release.ReleaseTypeAgent, channel, agentInfo.Os, arch, sr)
			if err != nil {
				errorMessage = err.Error()
				break
			}

			updateRequest := openuem_nats.OpenUEMUpdateRequest{}
			updateRequest.DownloadFrom = releaseToBeApplied.FileURL
			updateRequest.DownloadHash = releaseToBeApplied.Checksum

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

			data, err := json.Marshal(updateRequest)
			if err != nil {
				errorMessage = err.Error()
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", "admin.update.agents.task_status_error", releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				errorMessage = i18n.T(c.Request().Context(), "nats.not_connected")
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", "nats.not_connected", releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if _, err := h.JetStream.Publish(context.Background(), "agent.update."+a, data); err != nil {
				errorMessage = i18n.T(c.Request().Context(), "admin.update.agents.cannot_send_request")
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", "admin.update.agents.cannot_send_request", releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_pending", i18n.T(c.Request().Context(), "admin.update.agents.task_update", releaseToBeApplied.Version), releaseToBeApplied.Version); err != nil {
				log.Println("[ERROR]: could not save update task info")
				continue
			}
		}

		if errorMessage == "" {
			successMessage = i18n.T(c.Request().Context(), "admin.update.agents.success")
		} else {
			errorMessage = i18n.T(c.Request().Context(), "admin.update.agents.some_errors_found")
		}
	}

	return h.ShowUpdateAgentList(c, r, successMessage, errorMessage)
}

func (h *Handler) UpdateAgentsConfirm(c echo.Context) error {
	version := c.FormValue("version")
	return RenderConfirm(c, partials.ConfirmUpdateAgents(version))
}

func (h *Handler) ShowUpdateAgentList(c echo.Context, r *openuem_ent.Release, successMessage, errorMessage string) error {
	var err error
	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Get filters values
	f := filters.UpdateAgentsFilter{}
	f.Hostname = c.FormValue("filterByHostname")
	f.TaskResult = c.FormValue("filterByTaskResult")
	f.SelectedRelease = c.FormValue("filterBySelectedRelease")

	allReleases, err := h.Model.GetAgentsReleases()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))

	}

	availableReleases := []string{}
	for _, r := range allReleases {
		availableReleases = append(availableReleases, r.Version)
	}

	filteredReleases := []string{}
	for index := range allReleases {
		value := c.FormValue(fmt.Sprintf("filterByRelease%d", index))
		if value != "" {
			filteredReleases = append(filteredReleases, value)
		}
	}
	f.Releases = filteredReleases

	availableTaskStatus := []string{"admin.update.agents.task_status_success", "admin.update.agents.task_status_pending", "admin.update.agents.task_status_error"}
	filteredTaskStatus := []string{}
	for index := range availableTaskStatus {
		value := c.FormValue(fmt.Sprintf("filterByTaskStatus%d", index))
		if value != "" {
			filteredTaskStatus = append(filteredTaskStatus, value)
		}
	}
	f.TaskStatus = filteredTaskStatus

	nSelectedItems := c.FormValue("filterBySelectedItems")
	f.SelectedItems, err = strconv.Atoi(nSelectedItems)
	if err != nil {
		f.SelectedItems = 0
	}

	tmpAllAgents := []string{}
	allAgents, err := h.Model.GetAllUpdateAgents(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for _, a := range allAgents {
		tmpAllAgents = append(tmpAllAgents, "\""+a.ID+"\"")
	}
	f.SelectedAllAgents = "[" + strings.Join(tmpAllAgents, ",") + "]"

	appliedTags, err := h.Model.GetAppliedTags()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	for _, tag := range appliedTags {
		if c.FormValue(fmt.Sprintf("filterByTag%d", tag.ID)) != "" {
			f.Tags = append(f.Tags, tag.ID)
		}
	}

	lastExecutionFrom := c.FormValue("filterByLastExecutionDateFrom")
	if lastExecutionFrom != "" {
		f.TaskLastExecutionFrom = lastExecutionFrom
	}

	lastExecutionTo := c.FormValue("filterByLastExecutionDateTo")
	if lastExecutionTo != "" {
		f.TaskLastExecutionTo = lastExecutionTo
	}

	tagId := c.FormValue("tagId")
	agentId := c.FormValue("agentId")
	if c.Request().Method == "DELETE" && tagId != "" && agentId != "" {
		err := h.Model.RemoveTagFromAgent(agentId, tagId)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	p.NItems, err = h.Model.CountAllUpdateAgents(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	agents, err := h.Model.GetUpdateAgentsByPage(p, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	settings, err := h.Model.GetGeneralSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	higherVersion, err := h.Model.GetHigherAgentReleaseInstalled()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	l := views.GetTranslatorForDates(c)

	agentsExists, err := h.Model.AgentsExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	return RenderView(c, admin_views.UpdateAgentsIndex(" | Update Agents", admin_views.UpdateAgents(c, p, f, h.SessionManager, l, agents, settings, r, higherVersion, allReleases, availableReleases, availableTaskStatus, appliedTags, refreshTime, successMessage, errorMessage, agentsExists, serversExists)))
}
