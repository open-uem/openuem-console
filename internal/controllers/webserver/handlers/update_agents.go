package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	openuem_ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/release"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) UpdateAgents(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

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

		for a := range strings.SplitSeq(agents, ",") {

			agentInfo, err := h.Model.GetAgentById(a, commonInfo)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			arch := ""
			switch agentInfo.Edges.Computer.ProcessorArch {
			case "x64", "x86_64":
				arch = "amd64"
			case "aarch64":
				arch = "arm64"
			}

			switch agentInfo.Os {
			case "debian", "ubuntu", "opensuse-leap", "linuxmint", "fedora", "manjaro", "arch", "almalinux", "rocky":
				agentInfo.Os = "linux"
			case "macOS":
				agentInfo.Os = "darwin"
				macArch := strings.TrimSpace(agentInfo.Edges.Computer.ProcessorArch)
				if macArch == "x86_64" {
					arch = "amd64"
				} else {
					arch = "arm64"
				}
			}

			releaseToBeApplied, err := h.Model.GetAgentsReleaseByType(release.ReleaseTypeAgent, channel, agentInfo.Os, arch, sr)
			if err != nil {
				log.Printf("[ERROR]: could not get release to be applied, reason: %v\n", err)
				errorMessage = err.Error()
				break
			}

			updateRequest := openuem_nats.OpenUEMUpdateRequest{}
			updateRequest.DownloadFrom = releaseToBeApplied.FileURL
			updateRequest.DownloadHash = releaseToBeApplied.Checksum
			updateRequest.Version = releaseToBeApplied.Version

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
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", "admin.update.agents.task_status_error", releaseToBeApplied.Version, commonInfo); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				errorMessage = i18n.T(c.Request().Context(), "nats.not_connected")
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", "nats.not_connected", releaseToBeApplied.Version, commonInfo); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if _, err := h.JetStream.Publish(context.Background(), "agent.update."+a, data); err != nil {
				errorMessage = i18n.T(c.Request().Context(), "admin.update.agents.cannot_send_request")
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", "admin.update.agents.cannot_send_request", releaseToBeApplied.Version, commonInfo); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_pending", i18n.T(c.Request().Context(), "admin.update.agents.task_update", releaseToBeApplied.Version), releaseToBeApplied.Version, commonInfo); err != nil {
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
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	version := c.FormValue("filterBySelectedRelease")
	return RenderConfirm(c, partials.ConfirmUpdateAgents(c, version, commonInfo))
}

func (h *Handler) ShowUpdateAgentList(c echo.Context, r *openuem_ent.Release, successMessage, errorMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

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
	availableReleases = append(availableReleases, allReleases...)

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
	allAgents, err := h.Model.GetAllUpdateAgents(f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for _, a := range allAgents {
		tmpAllAgents = append(tmpAllAgents, "\""+a.ID+"\"")
	}
	f.SelectedAllAgents = "[" + strings.Join(tmpAllAgents, ",") + "]"

	appliedTags, err := h.Model.GetAppliedTags(commonInfo)
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
		err := h.Model.RemoveTagFromAgent(agentId, tagId, commonInfo)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	p.NItems, err = h.Model.CountAllUpdateAgents(f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	agents, err := h.Model.GetUpdateAgentsByPage(p, f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	settings, err := h.Model.GetGeneralSettings(commonInfo.TenantID)
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

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.UpdateAgentsIndex(" | Update Agents", admin_views.UpdateAgents(c, p, f, agents, settings, r, higherVersion, allReleases, availableReleases, availableTaskStatus, appliedTags, refreshTime, successMessage, errorMessage, agentsExists, serversExists, commonInfo, h.GetAdminTenantName(commonInfo)), commonInfo))
}
