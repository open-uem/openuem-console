package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/google/uuid"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/linde12/gowol"
	"github.com/microcosm-cc/bluemonday"
	openuem_ent "github.com/open-uem/ent"
	"github.com/open-uem/nats"
	openuem_nats "github.com/open-uem/nats"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	models "github.com/open-uem/openuem-console/internal/models/winget"
	"github.com/open-uem/openuem-console/internal/views"
	"github.com/open-uem/openuem-console/internal/views/computers_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/utils"
)

func (h *Handler) Computer(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentComputerInfo(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Computer(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) OperatingSystem(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentOSInfo(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}

	l := views.GetTranslatorForDates(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.OperatingSystem(c, p, h.SessionManager, h.Version, latestServerRelease.Version, l, agent, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) NetworkAdapters(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentNetworkAdaptersInfo(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.NetworkAdapters(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) Printers(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentPrintersInfo(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Printers(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) LogicalDisks(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentLogicalDisksInfo(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.LogicalDisks(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) Shares(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentSharesInfo(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Shares(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) Monitors(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentMonitorsInfo(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Monitors(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) Apps(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	a, err := h.Model.GetAgentById(agentId)
	if err != nil {
		log.Fatalf("[FATAL]: an error occurred querying agent: %v", err)
	}

	// Get filters
	f, err := h.GetSoftwareFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	apps, err := h.Model.GetAgentAppsByPage(agentId, p, *f)
	if err != nil {
		log.Fatalf("[FATAL]: an error occurred querying apps for agent: %v", err)
	}

	p.NItems, err = h.Model.CountAgentApps(agentId, *f)
	if err != nil {
		log.Fatalf("[FATAL]: an error occurred querying apps for agent: %v", err)
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Apps(c, p, *f, h.SessionManager, h.Version, latestServerRelease.Version, a, apps, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) RemoteAssistance(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.RemoteAssistance(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) ComputersList(c echo.Context, successMessage string, comesFromDialog bool) error {
	var err error

	currentPage := c.FormValue("page")
	pageSize := c.FormValue("pageSize")
	sortBy := c.FormValue("sortBy")
	sortOrder := c.FormValue("sortOrder")
	currentSortBy := c.FormValue("currentSortBy")

	p := partials.NewPaginationAndSort()

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			currentPage = "1"
			pageSize = u.Query().Get("pageSize")
			sortBy = u.Query().Get("sortBy")
			sortOrder = u.Query().Get("sortOrder")
			currentSortBy = u.Query().Get("currentSortBy")
		}
	}

	p.GetPaginationAndSortParams(currentPage, pageSize, sortBy, sortOrder, currentSortBy)

	// Get filters values
	f := filters.AgentFilter{}

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			f.Hostname = u.Query().Get("filterByHostname")
		}
	} else {
		f.Hostname = c.FormValue("filterByHostname")
	}

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			f.Username = u.Query().Get("filterByUsername")
		}
	} else {
		f.Username = c.FormValue("filterByUsername")
	}

	availableOSes, err := h.Model.GetAgentsUsedOSes()
	if err != nil {
		return err
	}
	filteredAgentOSes := []string{}
	for index := range availableOSes {
		if comesFromDialog {
			u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
			if err == nil {
				value := u.Query().Get(fmt.Sprintf("filterByAgentOS%d", index))
				if value != "" {
					filteredAgentOSes = append(filteredAgentOSes, value)
				}
			}
		} else {
			value := c.FormValue(fmt.Sprintf("filterByAgentOS%d", index))
			if value != "" {
				filteredAgentOSes = append(filteredAgentOSes, value)
			}
		}

	}
	f.AgentOSVersions = filteredAgentOSes

	versions, err := h.Model.GetOSVersions(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	filteredVersions := []string{}
	for index := range versions {
		if comesFromDialog {
			u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
			if err == nil {
				value := u.Query().Get(fmt.Sprintf("filterByOSVersion%d", index))
				if value != "" {
					filteredVersions = append(filteredVersions, value)
				}
			}
		} else {
			value := c.FormValue(fmt.Sprintf("filterByOSVersion%d", index))
			if value != "" {
				filteredVersions = append(filteredVersions, value)
			}
		}
	}
	f.OSVersions = filteredVersions

	filteredComputerManufacturers := []string{}
	vendors, err := h.Model.GetComputerManufacturers()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for index := range vendors {
		if comesFromDialog {
			u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
			if err == nil {
				value := u.Query().Get(fmt.Sprintf("filterByComputerManufacturer%d", index))
				if value != "" {
					filteredComputerManufacturers = append(filteredComputerManufacturers, value)
				}
			}
		} else {
			value := c.FormValue(fmt.Sprintf("filterByComputerManufacturer%d", index))
			if value != "" {
				filteredComputerManufacturers = append(filteredComputerManufacturers, value)
			}
		}
	}
	f.ComputerManufacturers = filteredComputerManufacturers

	filteredComputerModels := []string{}
	models, err := h.Model.GetComputerModels(f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	for index := range models {
		if comesFromDialog {
			u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
			if err == nil {
				value := u.Query().Get(fmt.Sprintf("filterByComputerModel%d", index))
				if value != "" {
					filteredComputerModels = append(filteredComputerModels, value)
				}
			}
		} else {
			value := c.FormValue(fmt.Sprintf("filterByComputerModel%d", index))
			if value != "" {
				filteredComputerModels = append(filteredComputerModels, value)
			}
		}
	}
	f.ComputerModels = filteredComputerModels

	filteredIsRemote := []string{}
	for index := range []string{"Remote", "Local"} {
		if comesFromDialog {
			u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
			if err == nil {
				value := u.Query().Get(fmt.Sprintf("filterByIsRemote%d", index))
				if value != "" {
					filteredIsRemote = append(filteredIsRemote, value)
				}
			}
		} else {
			value := c.FormValue(fmt.Sprintf("filterByIsRemote%d", index))
			if value != "" {
				filteredIsRemote = append(filteredIsRemote, value)
			}
		}
	}
	f.IsRemote = filteredIsRemote

	if c.FormValue("selectedApp") != "" {
		f.WithApplication = c.FormValue("selectedApp")
	}

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			f.WithApplication = u.Query().Get("filterByApplication")
		}
	} else {
		if c.FormValue("filterByApplication") != "" {
			f.WithApplication = c.FormValue("filterByApplication")
		}
	}

	tags, err := h.Model.GetAllTags()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	for _, tag := range tags {
		if comesFromDialog {
			u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
			if err == nil {
				if u.Query().Get(fmt.Sprintf("filterByTag%d", tag.ID)) != "" {
					f.Tags = append(f.Tags, tag.ID)
				}
			}
		} else {
			if c.FormValue(fmt.Sprintf("filterByTag%d", tag.ID)) != "" {
				f.Tags = append(f.Tags, tag.ID)
			}
		}
	}

	tagId := c.FormValue("tagId")
	agentId := c.FormValue("agentId")
	if c.Request().Method == "POST" && tagId != "" && agentId != "" {
		err := h.Model.AddTagToAgent(agentId, tagId)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	if c.Request().Method == "DELETE" && tagId != "" && agentId != "" {
		err := h.Model.RemoveTagFromAgent(agentId, tagId)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	computers, err := h.Model.GetComputersByPage(p, f)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	p.NItems, err = h.Model.CountAllComputers(f)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	l := views.GetTranslatorForDates(c)

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	if comesFromDialog {
		currentUrl := c.Request().Header.Get("Hx-Current-Url")
		if currentUrl != "" {
			if u, err := url.Parse(currentUrl); err == nil {
				q := u.Query()
				q.Del("page")
				q.Add("page", "1")
				u.RawQuery = q.Encode()
				return RenderViewWithReplaceUrl(c, computers_views.InventoryIndex("| Inventory", computers_views.Computers(c, p, f, h.SessionManager, l, h.Version, latestServerRelease.Version, computers, versions, vendors, models, tags, availableOSes, refreshTime, successMessage, detectRemoteAgents)), u)
			}
		}
	}

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Computers(c, p, f, h.SessionManager, l, h.Version, latestServerRelease.Version, computers, versions, vendors, models, tags, availableOSes, refreshTime, successMessage, detectRemoteAgents)))
}

func (h *Handler) ComputerDeploy(c echo.Context, successMessage string) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderError(c, partials.ErrorMessage("an error occurred getting uuid param", false))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	confirmDelete := c.QueryParam("delete") != ""

	deployments, err := h.Model.GetDeploymentsForAgent(agentId, p)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountDeploymentsForAgent(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	l := views.GetTranslatorForDates(c)

	if c.Request().Method == "POST" {
		return RenderView(c, computers_views.DeploymentsTable(c, p, l, agentId, deployments))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	return RenderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.ComputerDeploy(c, p, h.SessionManager, l, h.Version, latestServerRelease.Version, agent, deployments, successMessage, confirmDelete, detectRemoteAgents, refreshTime)))
}

func (h *Handler) ComputerDeploySearchPackagesInstall(c echo.Context) error {
	var f filters.DeployPackageFilter
	var packages []nats.SoftwarePackage

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_found"), false))
	}

	search := c.FormValue("filterByAppName")

	if search == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.search_empty_error"), true))
	}

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	if agent.Os == "windows" {
		f = filters.DeployPackageFilter{Sources: []string{"winget"}}
		useWinget, err := h.Model.GetDefaultUseWinget()
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.could_not_get_winget_use"), true))
		}

		if !useWinget {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.use_winget_is_false"), true))
		}

	} else {
		f = filters.DeployPackageFilter{Sources: []string{"flatpak"}}
		useFlatpak, err := h.Model.GetDefaultUseFlatpak()
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.could_not_get_flatpak_use"), true))
		}

		if !useFlatpak {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.use_flatpak_is_false"), true))
		}
	}

	packages, err = models.SearchPackages(search, p, h.CommonFolder, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.could_not_search_packages", err.Error()), true))
	}

	p.NItems, err = models.CountPackages(search, h.CommonFolder, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.could_not_count_packages", err.Error()), true))
	}

	return RenderView(c, computers_views.SearchPacketResult(c, agentId, packages, p))

}

func (h *Handler) ComputerDeployInstall(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	packageId := c.FormValue("filterByPackageId")
	packageName := c.FormValue("filterByPackageName")

	if packageId == "" || packageName == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.deploy_empty_values"), true))
	}

	alreadyInstalled, err := h.Model.DeploymentAlreadyInstalled(agentId, packageId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if alreadyInstalled {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.already_deployed"), true))
	}

	deploymentFailed, err := h.Model.DeploymentFailed(agentId, packageId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	action := openuem_nats.DeployAction{}
	action.AgentId = agentId
	action.PackageId = packageId
	action.PackageName = packageName
	// action.Repository = "winget"
	action.Action = "install"

	data, err := json.Marshal(action)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	err = h.NATSConnection.Publish("agent.installpackage."+agentId, data)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if err := h.Model.SaveDeployInfo(&action, deploymentFailed); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	c.Request().Method = "GET"
	return h.ComputerDeploy(c, i18n.T(c.Request().Context(), "agents.deploy_success"))
}

func (h *Handler) ComputerDeployUpdate(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	packageId := c.FormValue("filterByPackageId")
	packageName := c.FormValue("filterByPackageName")

	if packageId == "" || packageName == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.deploy_empty_values"), true))
	}

	action := openuem_nats.DeployAction{}
	action.AgentId = agentId
	action.PackageId = packageId
	action.PackageName = packageName
	// action.Repository = "winget"
	action.Action = "update"

	data, err := json.Marshal(action)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	err = h.NATSConnection.Publish("agent.updatepackage."+agentId, data)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	deploymentFailed, err := h.Model.DeploymentFailed(agentId, packageId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if err := h.Model.SaveDeployInfo(&action, deploymentFailed); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	c.Request().Method = "GET"
	return h.ComputerDeploy(c, i18n.T(c.Request().Context(), "agents.update_success"))
}

func (h *Handler) ComputerDeployUninstall(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	packageId := c.FormValue("filterByPackageId")
	packageName := c.FormValue("filterByPackageName")

	if packageId == "" || packageName == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.deploy_empty_values"), true))
	}

	// If the package hasn't been installed and the previous action was a failure
	d, err := h.Model.GetDeployment(agentId, packageId)
	if err == nil && d.Failed && d.Installed.IsZero() {
		if err := h.Model.RemoveDeployment(d.ID); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_remove_deployment"), true))
		}
		c.Request().Method = "GET"
		return h.ComputerDeploy(c, i18n.T(c.Request().Context(), "agents.deployment_removed"))
	}

	action := openuem_nats.DeployAction{}
	action.AgentId = agentId
	action.PackageId = packageId
	action.PackageName = packageName
	// action.Repository = "winget"
	action.Action = "uninstall"

	data, err := json.Marshal(action)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	err = h.NATSConnection.Publish("agent.uninstallpackage."+agentId, data)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	deploymentFailed, err := h.Model.DeploymentFailed(agentId, packageId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if err := h.Model.SaveDeployInfo(&action, deploymentFailed); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	c.Request().Method = "GET"
	return h.ComputerDeploy(c, i18n.T(c.Request().Context(), "agents.uninstall_success"))
}

func (h *Handler) PowerManagement(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	if c.Request().Method == "GET" {
		confirmDelete := c.QueryParam("delete") != ""
		p := partials.PaginationAndSort{}
		return RenderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.PowerManagement(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, confirmDelete, detectRemoteAgents)))
	}

	action := c.Param("action")
	if action == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_power_action"), false))
	}

	switch action {
	case "wol":
		mac := c.FormValue("MACAddress")
		if _, err := net.ParseMAC(mac); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		packet, err := gowol.NewMagicPacket(mac)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		// send wol to broadcast
		if err := packet.Send("255.255.255.255"); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.wol_success")))
	case "off":
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		action := openuem_nats.RebootOrRestart{}
		var whenTime time.Time
		when := c.FormValue("when")
		if when != "" {
			whenTime, err = time.ParseInLocation("2006-01-02T15:04", when, time.Local)
			if err != nil {
				log.Println("[INFO]: could not parse scheduled time as 24h time")
				whenTime, err = time.Parse("2006-01-02T15:04PM", when)
				if err != nil {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_parse_action_time"), false))
				}
			}
			action.Date = whenTime
		}

		data, err := json.Marshal(action)
		if err != nil {
			log.Printf("[ERROR]: could not marshall the Power Off request, reason: %v\n", err)
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.poweroff_could_not_marshal"), false))
		}

		if _, err := h.NATSConnection.Request("agent.poweroff."+agentId, data, time.Duration(h.NATSTimeout)*time.Second); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.request_error", err.Error()), true))
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.poweroff_success")))
	case "reboot":
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		action := openuem_nats.RebootOrRestart{}
		var whenTime time.Time
		when := c.FormValue("when")
		if when != "" {
			whenTime, err = time.ParseInLocation("2006-01-02T15:04", when, time.Local)
			if err != nil {
				log.Println("[INFO]: could not parse scheduled time as 24h time")
				whenTime, err = time.Parse("2006-01-02T15:04PM", when)
				if err != nil {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_parse_action_time"), false))
				}
			}
			action.Date = whenTime
		}

		data, err := json.Marshal(action)
		if err != nil {
			log.Printf("[ERROR]: could not marshall the Reboot request, reason: %v\n", err)
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.reboot_could_not_marshal"), false))
		}

		if _, err := h.NATSConnection.Request("agent.reboot."+agentId, data, time.Duration(h.NATSTimeout)*time.Second); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.request_error", err.Error()), true))
		}

		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.reboot_success")))
	default:
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_allowed_power_action"), false))
	}
}

func (h *Handler) ComputerMetadata(c echo.Context) error {
	var data []*openuem_ent.Metadata
	successMessage := ""

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderError(c, partials.ErrorMessage("an error occurred getting uuid param", false))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	confirmDelete := c.QueryParam("delete") != ""

	data, err = h.Model.GetMetadataForAgent(agentId, p)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	orgMetadata, err := h.Model.GetAllOrgMetadata()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllOrgMetadata()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		orgMetadataId := c.FormValue("orgMetadataId")
		name := c.FormValue("name")
		value := c.FormValue("value")

		id, err := strconv.Atoi(orgMetadataId)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		if orgMetadataId != "" && name != "" {
			acceptedMetadata := []int{}
			for _, data := range orgMetadata {
				acceptedMetadata = append(acceptedMetadata, data.ID)
			}

			found := false
			for _, item := range acceptedMetadata {
				if item == id {
					found = true
					break
				}
			}

			if !found {
				return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%s is not an accepted metadata", name), false))
			}

			if err := h.Model.SaveMetadata(agentId, id, value); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			data, err = h.Model.GetMetadataForAgent(agentId, p)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			successMessage = i18n.T(c.Request().Context(), "agents.metadata_save_success")
		}
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	return RenderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.ComputerMetadata(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, data, orgMetadata, confirmDelete, detectRemoteAgents, successMessage)))
}

func (h *Handler) Notes(c echo.Context) error {
	agentId := c.Param("uuid")

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error occurred getting uuid param", "Computer", "/computers", h.SessionManager, h.Version, latestServerRelease.Version)))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		notes := c.FormValue("markdown")
		if err := h.Model.SaveNotes(agentId, notes); err != nil {
			return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "notes.error", err.Error())))
		}
		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "notes.updated")))
	}

	maybeUnsafeHTML := markdown.ToHTML([]byte(agent.Notes), nil, nil)
	renderedMarkdown := string(bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML))

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Notes(c, p, h.SessionManager, h.Version, latestServerRelease.Version, agent, agent.Notes, renderedMarkdown, confirmDelete, detectRemoteAgents)))
}

func (h *Handler) ComputerConfirmDelete(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return h.ListAgents(c, "", "an error occurred getting uuid param", false)
	}

	err := h.Model.DeleteAgent(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}

	return h.ComputersList(c, i18n.T(c.Request().Context(), "computers.deleted"), true)
}

func (h *Handler) ComputerStartVNC(c echo.Context) error {
	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if c.Request().Method == "POST" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		// Check if PIN is optional or not
		requestPIN, err := h.Model.GetDefaultRequestVNCPIN()
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.request_pin_could_not_be_read"), false))
		}

		// Create new random PIN
		pin, err := utils.GenerateRandomPIN()
		if err != nil {
			log.Printf("[ERROR]: could not generate random PIN, reason: %v\n", err)
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.vnc_pin_not_generated"), false))
		}

		vncConn := openuem_nats.VNCConnection{}
		vncConn.NotifyUser = requestPIN
		vncConn.PIN = pin

		data, err := json.Marshal(vncConn)
		if err != nil {
			log.Printf("[ERROR]: could not marshall VNC connection data, reason: %v\n", err)
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.vnc_could_not_marshal"), false))
		}

		if _, err := h.NATSConnection.Request("agent.startvnc."+agentId, data, time.Duration(h.NATSTimeout)*time.Second); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}

		if strings.Contains(agent.Vnc, "RDP") {
			return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.RemoteDesktop(agent, h.Domain, true, requestPIN, pin, h.SessionManager, h.Version, latestServerRelease.Version)))
		} else {
			return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.VNC(agent, h.Domain, true, requestPIN, pin, h.SessionManager, h.Version, latestServerRelease.Version)))
		}
	}

	if strings.Contains(agent.Vnc, "RDP") {
		return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.RemoteDesktop(agent, h.Domain, false, false, "", h.SessionManager, h.Version, latestServerRelease.Version)))
	}
	return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.VNC(agent, h.Domain, false, false, "", h.SessionManager, h.Version, latestServerRelease.Version)))
}

func (h *Handler) ComputerStopVNC(c echo.Context) error {
	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	if _, err := h.NATSConnection.Request("agent.stopvnc."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responder"), false))
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.VNC(agent, h.Domain, false, false, "", h.SessionManager, h.Version, latestServerRelease.Version)))
}

func (h *Handler) GenerateRDPFile(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage("an error occurred getting uuid param", false))
	}

	fileName := uuid.NewString() + ".rdp"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_found"), false))
	}

	f, err := os.Create(dstPath)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_generate_rdp_file"), false))
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("[ERROR]: could not close RDP file")
		}
	}()

	if _, err := f.WriteString(fmt.Sprintf("full address:s:%s\n", agent.IP)); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	if _, err := f.WriteString("username:s:openuem\n"); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	if _, err := f.WriteString("audiocapturemode:i:0\n"); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	if _, err := f.WriteString("audiomode:i:2\n"); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}
