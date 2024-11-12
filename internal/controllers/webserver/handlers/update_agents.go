package handlers

import (
	"encoding/json"
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

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Get filters values
	f := filters.AgentFilter{}
	f.Hostname = c.FormValue("filterByHostname")
	f.SelectedStatus = c.FormValue("filterBySelectedStatus")
	if f.SelectedStatus == "" {
		f.SelectedStatus = "none"
	}

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

	p.NItems, err = h.Model.CountAllAgents(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		// TODO agent selection, right now is hardcoded for my test
		agentId := "1ed69271-79c6-4a15-b2d0-c13763170ddf"

		updateRequest := openuem_nats.OpenUEMUpdateRequest{}
		updateRequest.DownloadFrom = version.FileURL
		updateRequest.DownloadHash = version.Checksum
		updateRequest.UpdateNow = true

		data, err := json.Marshal(updateRequest)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		if _, err := h.NATSConnection.Request("agentupdate."+agentId, data, 10*time.Second); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "admin.update.agents.success")))
	}

	settings, err := h.Model.GetGeneralSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	higherVersion, err := h.Model.GetHigherAgentVersionInstalled()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.UpdateAgentsIndex(" | Update Agents", admin_views.UpdateAgents(c, p, f, agents, settings, version, higherVersion)))
}

func (h *Handler) UpdateAgentsConfirmSelected(c echo.Context) error {
	version := c.FormValue("version")

	/* f := filters.AgentFilter{}


	countAgents, err := h.Model.CountAllAgents(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	selectedAgents := c.FormValue("filterBySelectedItems") */

	return RenderConfirm(c, partials.ConfirmUpdateAgents(version, false))
}

func (h *Handler) UpdateAgentsConfirmAll(c echo.Context) error {
	version := c.FormValue("version")
	return RenderConfirm(c, partials.ConfirmUpdateAgents(version, true))
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
