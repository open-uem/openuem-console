package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"strconv"
	"strings"
	"time"

	model "github.com/doncicuto/openuem-console/internal/models/components"
	"github.com/doncicuto/openuem-console/internal/views"
	"github.com/doncicuto/openuem-console/internal/views/admin_views"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/component"
	"github.com/doncicuto/openuem_ent/release"
	"github.com/doncicuto/openuem_nats"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) UpdateComponents(c echo.Context) error {
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
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", errorMessage, releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				errorMessage = i18n.T(c.Request().Context(), "nats.not_connected")
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", errorMessage, releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if _, err := h.JetStream.Publish(context.Background(), "agentupdate."+a, data); err != nil {
				errorMessage = i18n.T(c.Request().Context(), "admin.update.agents.cannot_send_request")
				if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_error", errorMessage, releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if err := h.Model.SaveAgentUpdateInfo(a, "admin.update.agents.task_status_pending", "admin.update.agents.task_update", releaseToBeApplied.Version); err != nil {
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

	return h.ShowUpdateComponentsList(c, r, successMessage, errorMessage)
}

func (h *Handler) UpdateComponentsConfirm(c echo.Context) error {
	version := c.FormValue("version")
	return RenderConfirm(c, partials.ConfirmUpdateAgents(version))
}

func (h *Handler) ShowUpdateComponentsList(c echo.Context, r *openuem_ent.Release, successMessage, errorMessage string) error {
	var err error
	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Get filters values
	f := filters.UpdateComponentsFilter{}
	f.Hostname = c.FormValue("filterByHostname")
	f.UpdateMessage = c.FormValue("filterByUpdateMessage")

	nSelectedItems := c.FormValue("filterBySelectedItems")
	f.SelectedItems, err = strconv.Atoi(nSelectedItems)
	if err != nil {
		f.SelectedItems = 0
	}

	tmpAllComponents := []string{}
	allUpdateComponents, err := h.Model.GetAllUpdateComponents(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for _, c := range allUpdateComponents {
		tmpAllComponents = append(tmpAllComponents, "\""+strconv.Itoa(c.ID)+"\"")
	}
	f.SelectedAllComponents = "[" + strings.Join(tmpAllComponents, ",") + "]"

	whenFrom := c.FormValue("filterByUpdateWhenDateFrom")
	if whenFrom != "" {
		f.UpdateWhenFrom = whenFrom
	}

	whenTo := c.FormValue("filterByUpdateWhenDateTo")
	if whenTo != "" {
		f.UpdateWhenTo = whenTo
	}

	allUpdateStatus := []string{
		component.UpdateStatusSuccess.String(),
		component.UpdateStatusPending.String(),
		component.UpdateStatusError.String(),
	}

	allComponents := []string{
		component.ComponentNats.String(),
		component.ComponentOcsp.String(),
		component.ComponentConsole.String(),
		component.ComponentAgentWorker.String(),
		component.ComponentCertManagerWorker.String(),
		component.ComponentNotificationWorker.String(),
		component.ComponentCertManager.String(),
	}

	filteredComponents := []string{}
	for index := range allComponents {
		value := c.FormValue(fmt.Sprintf("filterByComponent%d", index))
		if value != "" {
			filteredComponents = append(filteredComponents, value)
		}
	}
	f.Components = filteredComponents

	allReleasesFromJson, err := model.GetServerReleases()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	allReleases := []string{}
	for key := range maps.Keys(allReleasesFromJson) {
		allReleases = append(allReleases, key)
	}

	appliedReleases, err := h.Model.GetAppliedReleases()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	filteredReleases := []string{}
	for index := range appliedReleases {
		value := c.FormValue(fmt.Sprintf("filterByRelease%d", index))
		if value != "" {
			filteredReleases = append(filteredReleases, value)
		}
	}
	f.Releases = filteredReleases

	filteredUpdateStatus := []string{}
	for index := range allUpdateStatus {
		value := c.FormValue(fmt.Sprintf("filterByUpdateStatus%d", index))
		if value != "" {
			filteredUpdateStatus = append(filteredUpdateStatus, value)
		}
	}
	f.UpdateStatus = filteredUpdateStatus

	p.NItems, err = h.Model.CountAllUpdateServers(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	components, err := h.Model.GetUpdateComponentsByPage(p, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	higherRelease, err := h.Model.GetHigherServerReleaseInstalled()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	latestServerRelease, err := model.GetLatestServerRelease()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	l := views.GetTranslatorForDates(c)

	return RenderView(c, admin_views.UpdateComponentsIndex(" | Update Components", admin_views.UpdateComponents(c, p, f, h.SessionManager, l, components, allComponents, higherRelease, latestServerRelease, appliedReleases, allReleases, allUpdateStatus, refreshTime, successMessage, errorMessage)))
}
