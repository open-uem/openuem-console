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
	"github.com/open-uem/ent/server"
	openuem_nats "github.com/open-uem/nats"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) UpdateServers(c echo.Context) error {
	var err error
	successMessage := ""
	errorMessage := ""

	// Get latest version
	channel, err := h.Model.GetDefaultUpdateChannel()
	if err != nil {
		log.Println("[ERROR]: could not get updates channel settings")
		channel = "stable"
	}

	r, err := h.Model.GetLatestServerRelease(channel)
	if err != nil {
		log.Println("[ERROR]: could not get latest version information")
	}

	if c.Request().Method == "POST" {
		servers := c.FormValue("servers")
		if servers == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "admin.update.servers.servers_cant_be_empty"), false))
		}

		sr := c.FormValue("filterBySelectedRelease")
		if sr == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "admin.update.servers.release_cant_be_empty"), false))
		}

		for s := range strings.SplitSeq(servers, ",") {

			serverId, err := strconv.Atoi(s)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			serverInfo, err := h.Model.GetServerById(serverId)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			releaseToBeApplied, err := h.Model.GetServersReleaseByType(release.ReleaseTypeServer, channel, serverInfo.Os, serverInfo.Arch, sr)
			if err != nil {
				errorMessage = err.Error()
				break
			}

			updateRequest := openuem_nats.OpenUEMUpdateRequest{}
			updateRequest.DownloadFrom = releaseToBeApplied.FileURL
			updateRequest.DownloadHash = releaseToBeApplied.Checksum
			updateRequest.Version = releaseToBeApplied.Version
			updateRequest.Channel = releaseToBeApplied.Channel

			if c.FormValue("update-server-date") == "" {
				updateRequest.UpdateNow = true
			} else {
				scheduledTime := c.FormValue("update-server-date")
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
				if err := h.Model.SaveServerUpdateInfo(serverId, server.UpdateStatusError, errorMessage, releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
				errorMessage = i18n.T(c.Request().Context(), "nats.not_connected")
				if err := h.Model.SaveServerUpdateInfo(serverId, server.UpdateStatusError, errorMessage, releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if _, err := h.JetStream.Publish(context.Background(), "server.update."+serverInfo.Hostname, data); err != nil {
				errorMessage = i18n.T(c.Request().Context(), "admin.update.servers.cannot_send_request")
				if err := h.Model.SaveServerUpdateInfo(serverId, server.UpdateStatusError, errorMessage, releaseToBeApplied.Version); err != nil {
					log.Println("[ERROR]: could not save update task info")
				}
				continue
			}

			if err := h.Model.SaveServerUpdateInfo(serverId, server.UpdateStatusPending, i18n.T(c.Request().Context(), "admin.update.servers.task_update", releaseToBeApplied.Version), releaseToBeApplied.Version); err != nil {
				log.Println("[ERROR]: could not save update task info")
				continue
			}
		}

		if errorMessage == "" {
			successMessage = i18n.T(c.Request().Context(), "admin.update.servers.success")
		} else {
			log.Println("[ERROR]:", errorMessage)
			errorMessage = i18n.T(c.Request().Context(), "admin.update.servers.some_errors_found") + ": " + errorMessage
		}
	}

	if c.Request().Method == "DELETE" {
		id := c.Param("serverId")
		serverId, err := strconv.Atoi(id)
		if err != nil {
			errorMessage = i18n.T(c.Request().Context(), "admin.update.servers.could_not_parse_server_id")
		}

		if err := h.Model.DeleteServer(serverId); err != nil {
			errorMessage = i18n.T(c.Request().Context(), "admin.update.servers.could_not_delete", err.Error())
		}

		successMessage = i18n.T(c.Request().Context(), "admin.update.servers.server_deleted")
	}

	return h.ShowUpdateServersList(c, r, successMessage, errorMessage)
}

func (h *Handler) DeleteServer(c echo.Context) error {
	return nil
}

func (h *Handler) UpdateServersConfirm(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	version := c.FormValue("filterBySelectedRelease")
	return RenderConfirm(c, partials.ConfirmUpdateServers(c, version, commonInfo))
}

func (h *Handler) DeleteServerConfirm(c echo.Context) error {
	server := c.Param("serverId")
	return RenderConfirm(c, partials.ConfirmDelete(c, i18n.T(c.Request().Context(), "admin.update.servers.confirm_delete"), "/admin/update-servers", fmt.Sprintf("/admin/update-servers/%s", server)))
}

func (h *Handler) ShowUpdateServersList(c echo.Context, r *openuem_ent.Release, successMessage, errorMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	// Get filters values
	f := filters.UpdateServersFilter{}
	f.Hostname = c.FormValue("filterByHostname")
	f.UpdateMessage = c.FormValue("filterByUpdateMessage")

	nSelectedItems := c.FormValue("filterBySelectedItems")
	f.SelectedItems, err = strconv.Atoi(nSelectedItems)
	if err != nil {
		f.SelectedItems = 0
	}

	tmpAllServers := []string{}
	allUpdateServers, err := h.Model.GetAllUpdateServers(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for _, c := range allUpdateServers {
		tmpAllServers = append(tmpAllServers, "\""+strconv.Itoa(c.ID)+"\"")
	}
	f.SelectedAllServers = "[" + strings.Join(tmpAllServers, ",") + "]"

	whenFrom := c.FormValue("filterByUpdateWhenDateFrom")
	if whenFrom != "" {
		f.UpdateWhenFrom = whenFrom
	}

	whenTo := c.FormValue("filterByUpdateWhenDateTo")
	if whenTo != "" {
		f.UpdateWhenTo = whenTo
	}

	allUpdateStatus := []string{
		server.UpdateStatusSuccess.String(),
		server.UpdateStatusPending.String(),
		server.UpdateStatusError.String(),
	}

	allReleases, err := h.Model.GetServerReleases()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
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

	Servers, err := h.Model.GetUpdateServersByPage(p, f)
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

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.UpdateServersIndex(" | Update Servers", admin_views.UpdateServers(c, p, f, Servers, []string{}, higherRelease, latestServerRelease, appliedReleases, allReleases, allUpdateStatus, refreshTime, successMessage, errorMessage, agentsExists, serversExists, commonInfo), commonInfo))
}
