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
	"slices"
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
	"github.com/open-uem/ent/task"
	openuem_nats "github.com/open-uem/nats"
	ansiblecfg "github.com/open-uem/openuem-ansible-config/ansible"
	models "github.com/open-uem/openuem-console/internal/models/winget"
	"github.com/open-uem/openuem-console/internal/views/computers_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/utils"
	"gopkg.in/yaml.v3"
)

func (h *Handler) Overview(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	successMessage := ""

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	if c.Request().Method == "POST" {
		description := c.FormValue("endpoint-description")
		endpointType := c.FormValue("endpoint-type")
		tenant := c.FormValue("tenant")
		site := c.FormValue("site")

		if description != "" {
			if err := h.Model.SaveEndpointDescription(agentId, description, commonInfo); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.overview_description_could_not_save", err.Error()), true))
			}
			successMessage = i18n.T(c.Request().Context(), "agents.overview_description_success")
		}

		if endpointType != "" {
			if !slices.Contains([]string{"DesktopPC", "Laptop", "Server", "Tablet", "VM", "Other"}, endpointType) {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.overview_endpoint_type_invalid"), true))
			}
			if err := h.Model.SaveEndpointType(agentId, endpointType, commonInfo); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.overview_endpoint_type_could_not_save"), true))
			}
			successMessage = i18n.T(c.Request().Context(), "agents.overview_endpoint_type_success")
		}

		if tenant != "" && site != "" {
			if err := h.Model.AssociateToTenantAndSite(agentId, tenant, site); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.overview_endpoint_type_could_not_save", err.Error()), true))
			}

			// Change URL with the new site and organization
			c.Response().Header().Set("HX-Replace-Url", fmt.Sprintf("/tenant/%s/site/%s/computers/%s/overview", tenant, site, agentId))
			commonInfo.TenantID = tenant
			commonInfo.SiteID = site
			tenantID, err := strconv.Atoi(tenant)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
			}
			commonInfo.Sites, err = h.Model.GetSites(tenantID)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_sites", err.Error()), true))
			}
			successMessage = i18n.T(c.Request().Context(), "agents.association_success")
		}
	}

	agent, err := h.Model.GetAgentOverviewById(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	sites := agent.Edges.Site
	if len(sites) == 0 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_associated_site"), true))
	}

	if len(sites) > 1 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.agent_cannot_associated_to_more_than_one_site"), true))
	}

	currentSite := sites[0]

	s, err := h.Model.GetSite(currentSite.ID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_site_info"), true))
	}

	if s.Edges.Tenant == nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_tenant"), true))
	}

	currentTenant := s.Edges.Tenant

	allTenants, err := h.Model.GetTenants()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_tenants"), true))
	}

	allSites, err := h.Model.GetSites(currentTenant.ID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_tenants"), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}

	higherVersion, err := h.Model.GetHigherAgentReleaseInstalled()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Overview(c, p, agent, higherVersion, confirmDelete, successMessage, commonInfo, currentTenant, currentSite, allTenants, allSites, netbird, offline), commonInfo))
}

func (h *Handler) Computer(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentComputerInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)
	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Computer(c, p, agent, confirmDelete, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) OperatingSystem(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentOSInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.OperatingSystem(c, p, agent, confirmDelete, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) NetworkAdapters(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	currentPage := c.FormValue("page")
	pageSize := c.FormValue("pageSize")
	sortBy := c.FormValue("sortBy")
	sortOrder := c.FormValue("sortOrder")
	currentSortBy := c.FormValue("currentSortBy")

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(currentPage, pageSize, sortBy, sortOrder, currentSortBy, itemsPerPage)

	agent, err := h.Model.GetAgentNetworkAdaptersInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	adapters, err := h.Model.NetworkAdaptersByPageInfo(agentId, commonInfo, p)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	p.NItems, err = h.Model.CountNetworkAdaptersByPageInfo(agentId, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: an error occurred counting apps for agent: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.NetworkAdapters(c, p, agent, adapters, confirmDelete, itemsPerPage, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) Printers(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	printers, err := h.Model.GetAgentPrintersInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Printers(c, p, agent, printers, confirmDelete, "", commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) LogicalDisks(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentLogicalDisksInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.LogicalDisks(c, p, agent, confirmDelete, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) PhysicalDisks(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentPhysicalDisksInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.PhysicalDisks(c, p, agent, confirmDelete, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) Shares(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentSharesInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Shares(c, p, agent, confirmDelete, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) Monitors(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentMonitorsInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Monitors(c, p, agent, confirmDelete, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) Apps(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	a, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	// Get filters
	f, err := h.GetSoftwareFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	apps, err := h.Model.GetAgentAppsByPage(agentId, p, *f, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: an error occurred querying apps for agent: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p.NItems, err = h.Model.CountAgentApps(agentId, *f, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: an error occurred counting apps for agent: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Apps(c, p, *f, a, apps, confirmDelete, itemsPerPage, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) RemoteAssistance(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	hasRustDeskSettings := h.Model.HasRustDeskSettings(tenantID)

	domain := h.Domain
	if len(agent.Edges.Site) == 1 && agent.Edges.Site[0].Domain != "" {
		domain = agent.Edges.Site[0].Domain
	}

	_, err = net.LookupIP(agent.Hostname + "." + domain)
	isHostResolvedByDNS := err == nil

	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.RemoteAssistance(c, p, agent, confirmDelete, hasRustDeskSettings, isHostResolvedByDNS, commonInfo, "", netbird, offline), commonInfo))
}

func (h *Handler) ComputersList(c echo.Context, successMessage string, comesFromDialog bool) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	currentPage := c.FormValue("page")
	pageSize := c.FormValue("pageSize")
	sortBy := c.FormValue("sortBy")
	sortOrder := c.FormValue("sortOrder")
	currentSortBy := c.FormValue("currentSortBy")

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)

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

	p.GetPaginationAndSortParams(currentPage, pageSize, sortBy, sortOrder, currentSortBy, itemsPerPage)

	// Get filters values
	f := filters.AgentFilter{}

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			f.Search = u.Query().Get("filterBySearch")
		}
	} else {
		if c.FormValue("filterBySearch") != "" {
			f.Search = c.FormValue("filterBySearch")
		}
	}

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			f.Nickname = u.Query().Get("filterByNickname")
		}
	} else {
		f.Nickname = c.FormValue("filterByNickname")
	}

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			f.Username = u.Query().Get("filterByUsername")
		}
	} else {
		f.Username = c.FormValue("filterByUsername")
	}

	availableOSes, err := h.Model.GetAgentsUsedOSes(commonInfo, f, false)
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

	versions, err := h.Model.GetOSVersions(f, commonInfo)
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
	vendors, err := h.Model.GetComputerManufacturers(commonInfo, f)
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
	models, err := h.Model.GetComputerModels(f, commonInfo)
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

	if c.FormValue("selectedPublisher") != "" {
		f.WithApplicationPublisher = c.FormValue("selectedPublisher")
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

	tags, err := h.Model.GetAllTags(commonInfo, f)
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
		err := h.Model.AddTagToAgent(agentId, tagId, commonInfo)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	if c.Request().Method == "DELETE" && tagId != "" && agentId != "" {
		err := h.Model.RemoveTagFromAgent(agentId, tagId, commonInfo)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	computers, err := h.Model.GetComputersByPage(p, f, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	p.NItems, err = h.Model.CountAllComputers(f, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	if comesFromDialog {
		currentUrl := c.Request().Header.Get("Hx-Current-Url")
		if currentUrl != "" {
			if u, err := url.Parse(currentUrl); err == nil {
				q := u.Query()
				q.Del("page")
				q.Add("page", "1")
				u.RawQuery = q.Encode()
				return RenderViewWithReplaceUrl(c, computers_views.InventoryIndex("| Inventory", computers_views.Computers(c, p, f, computers, versions, vendors, models, tags, availableOSes, refreshTime, itemsPerPage, successMessage, commonInfo), commonInfo), u)
			}
		}
	}

	// Use filters to get lists of values for the filter dialogs
	availableOSes, err = h.Model.GetAgentsUsedOSes(commonInfo, f, false)
	if err != nil {
		return err
	}

	tags, err = h.Model.GetAllTags(commonInfo, f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	versions, err = h.Model.GetOSVersions(f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Computers(c, p, f, computers, versions, vendors, models, tags, availableOSes, refreshTime, itemsPerPage, successMessage, commonInfo), commonInfo))
}

func (h *Handler) ComputerDeploy(c echo.Context, successMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderError(c, partials.ErrorMessage("an error occurred getting uuid param", false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	confirmDelete := c.QueryParam("delete") != ""

	deployments, err := h.Model.GetDeploymentsForAgent(agentId, p, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountDeploymentsForAgent(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		return RenderView(c, computers_views.DeploymentsTable(c, p, agentId, deployments, itemsPerPage, commonInfo))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.ComputerDeploy(c, p, agent, deployments, successMessage, confirmDelete, refreshTime, itemsPerPage, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) ComputerDeploySearchPackagesInstall(c echo.Context) error {
	var f filters.DeployPackageFilter
	var packages []openuem_nats.SoftwarePackage

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
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

	switch agent.Os {
	case "windows":
		f = filters.DeployPackageFilter{Sources: []string{"winget"}}
		useWinget, err := h.Model.GetDefaultUseWinget(commonInfo.TenantID)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.could_not_get_winget_use"), true))
		}

		if !useWinget {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.use_winget_is_false"), true))
		}
	case "macos", "macOS":
		f = filters.DeployPackageFilter{Sources: []string{"brew"}}
		useBrew, err := h.Model.GetDefaultUseBrew(commonInfo.TenantID)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.could_not_get_brew_use"), true))
		}

		if !useBrew {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.use_brew_is_false"), true))
		}
	default:
		f = filters.DeployPackageFilter{Sources: []string{"flatpak"}}
		useFlatpak, err := h.Model.GetDefaultUseFlatpak(commonInfo.TenantID)
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

	return RenderView(c, computers_views.SearchPacketResult(c, agentId, packages, p, itemsPerPage, commonInfo))

}

func (h *Handler) ComputerDeployInstall(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	packageId := c.FormValue("filterByPackageId")
	packageName := c.FormValue("filterByPackageName")

	if packageId == "" || packageName == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.deploy_empty_values"), true))
	}

	alreadyInstalled, err := h.Model.DeploymentAlreadyInstalled(agentId, packageId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if alreadyInstalled {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.already_deployed"), true))
	}

	deploymentFailed, err := h.Model.DeploymentFailed(agentId, packageId, commonInfo)
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

	if err := h.Model.SaveDeployInfo(&action, deploymentFailed, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	c.Request().Method = "GET"
	return h.ComputerDeploy(c, i18n.T(c.Request().Context(), "agents.deploy_success"))
}

func (h *Handler) ComputerDeployUpdate(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

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

	deploymentFailed, err := h.Model.DeploymentFailed(agentId, packageId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if err := h.Model.SaveDeployInfo(&action, deploymentFailed, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	c.Request().Method = "GET"
	return h.ComputerDeploy(c, i18n.T(c.Request().Context(), "agents.update_success"))
}

func (h *Handler) ComputerDeployUninstall(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

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
	d, err := h.Model.GetDeployment(agentId, packageId, commonInfo)
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

	deploymentFailed, err := h.Model.DeploymentFailed(agentId, packageId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if err := h.Model.SaveDeployInfo(&action, deploymentFailed, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	c.Request().Method = "GET"
	return h.ComputerDeploy(c, i18n.T(c.Request().Context(), "agents.uninstall_success"))
}

func (h *Handler) PowerManagement(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	if c.Request().Method == "GET" {
		confirmDelete := c.QueryParam("delete") != ""
		p := partials.PaginationAndSort{}

		tenantID, err := strconv.Atoi(commonInfo.TenantID)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
		}
		settings, err := h.Model.GetNetbirdSettings(tenantID)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
		}
		netbird := settings.AccessToken != ""

		offline := h.IsAgentOffline(c)

		return RenderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.PowerManagement(c, p, agent, confirmDelete, commonInfo, netbird, offline), commonInfo))
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

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	successMessage := ""

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderError(c, partials.ErrorMessage("an error occurred getting uuid param", false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	confirmDelete := c.QueryParam("delete") != ""

	data, err = h.Model.GetMetadataForAgent(agentId, p, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	orgMetadata, err := h.Model.GetAllOrgMetadata(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllOrgMetadata(commonInfo)
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

			if !slices.Contains(acceptedMetadata, id) {
				return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%s is not an accepted metadata", name), false))
			}

			if err := h.Model.SaveMetadata(agentId, id, value); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			data, err = h.Model.GetMetadataForAgent(agentId, p, commonInfo)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}

			successMessage = i18n.T(c.Request().Context(), "agents.metadata_save_success")
		}
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.ComputerMetadata(c, p, agent, data, orgMetadata, confirmDelete, successMessage, itemsPerPage, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) Notes(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	if c.Request().Method == "POST" {
		notes := c.FormValue("markdown")
		if err := h.Model.SaveNotes(agentId, notes, commonInfo); err != nil {
			return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "notes.error", err.Error())))
		}
		return RenderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "notes.updated")))
	}

	maybeUnsafeHTML := markdown.ToHTML([]byte(agent.Notes), nil, nil)
	renderedMarkdown := string(bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML))

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Notes(c, p, agent, agent.Notes, renderedMarkdown, confirmDelete, commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) ComputerConfirmDelete(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return h.ListAgents(c, "", "an error occurred getting uuid param", true)
	}

	if err := h.Model.DeleteAgent(agentId, commonInfo); err != nil {
		return h.ListAgents(c, "", err.Error(), true)
	}

	return h.ComputersList(c, i18n.T(c.Request().Context(), "computers.deleted"), true)
}

func (h *Handler) ComputerStartVNC(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	domain := h.Domain
	if len(agent.Edges.Site) == 1 && agent.Edges.Site[0].Domain != "" {
		domain = agent.Edges.Site[0].Domain
	}

	if c.Request().Method == "POST" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		// Check if PIN is optional or not
		requestPIN, err := h.Model.GetDefaultRequestVNCPIN(commonInfo.TenantID)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.request_pin_could_not_be_read"), false))
		}

		// Create new random PIN
		pinLength := 6
		if agent.Os == "macOS" {
			pinLength = 8
		}
		pin, err := utils.GenerateRandomPIN(pinLength)
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
			return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.RemoteDesktop(c, agent, domain, true, requestPIN, pin, commonInfo), commonInfo))
		} else {
			return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.VNC(c, agent, domain, true, requestPIN, pin, commonInfo), commonInfo))
		}
	}

	if strings.Contains(agent.Vnc, "RDP") {
		return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.RemoteDesktop(c, agent, domain, false, false, "", commonInfo), commonInfo))
	}
	return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.VNC(c, agent, domain, false, false, "", commonInfo), commonInfo))
}

func (h *Handler) ComputerStartRustDesk(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}

	settings, err := h.Model.GetRustDeskSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_get_rustdesk_settings", err.Error()), true))
	}

	return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.RustDesk(c, agent, settings, commonInfo), commonInfo))
}

func (h *Handler) ComputerStopVNC(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	domain := h.Domain
	if len(agent.Edges.Site) == 1 && agent.Edges.Site[0].Domain != "" {
		domain = agent.Edges.Site[0].Domain
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	if _, err := h.NATSConnection.Request("agent.stopvnc."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responder"), false))
	}

	return RenderView(c, computers_views.InventoryIndex("| Computers", computers_views.VNC(c, agent, domain, false, false, "", commonInfo), commonInfo))
}

func (h *Handler) GenerateRDPFile(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage("an error occurred getting uuid param", false))
	}

	fileName := uuid.NewString() + ".rdp"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
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

func (h *Handler) SetDefaultPrinter(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_get_common_info"), false))
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	printer := c.Param("printer")
	if printer == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.printer_name"), false))
	}

	printerName, err := url.QueryUnescape(printer)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_decode_printer"), false))
	}

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	msg, err := h.NATSConnection.Request("agent.defaultprinter."+agentId, []byte(printerName), time.Duration(h.NATSTimeout)*time.Second)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.request_error", err.Error()), true))
	}

	if string(msg.Data) != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.printer_could_not_set_as_default", string(msg.Data)), false))
	}

	if err := h.Model.SetDefaultPrinter(agentId, printerName, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.printer_could_not_set_as_default", err.Error()), false))
	}

	printers, err := h.Model.GetAgentPrintersInfo(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Printers(c, p, agent, printers, confirmDelete, i18n.T(c.Request().Context(), "agents.printer_has_been_set_as_default"), commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) RemovePrinter(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	printer := c.Param("printer")
	if printer == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.printer_name"), false))
	}

	printerName, err := url.QueryUnescape(printer)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	msg, err := h.NATSConnection.Request("agent.removeprinter."+agentId, []byte(printerName), time.Duration(h.NATSTimeout)*time.Second)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.request_error", err.Error()), true))
	}

	if string(msg.Data) != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.printer_could_not_be_removed", string(msg.Data)), false))
	}

	if err := h.Model.RemovePrinter(agentId, printerName, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.printer_could_not_be_removed", err.Error()), false))
	}

	printers, err := h.Model.GetAgentPrintersInfo(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	confirmDelete := c.QueryParam("delete") != ""

	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}

	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Printers(c, p, agent, printers, confirmDelete, i18n.T(c.Request().Context(), "agents.printer_has_been_removed"), commonInfo, netbird, offline), commonInfo))
}

func (h *Handler) GetDropdownSites(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	tenantID, err := strconv.Atoi(c.FormValue("tenant"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_convert_to_int"), false))
	}

	sites, err := h.Model.GetSites(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_get_sites"), false))
	}

	if commonInfo.SiteID == "-1" {
		commonInfo.SiteID = c.FormValue("site")
	}

	return RenderView(c, computers_views.SitesDropdown(c, agentId, sites, commonInfo))

}

func (h *Handler) AgentStatus(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_get_common_info"), false))
	}

	offline := h.IsAgentOffline(c)
	url := fmt.Sprintf("/computers/%s/status", agentId)

	return RenderView(c, partials.AgentStatus(commonInfo, url, offline))
}

func (h *Handler) IsAgentOffline(c echo.Context) bool {
	agentId := c.Param("uuid")
	if agentId == "" {
		return true
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.ping.%s", agentId), nil, 1*time.Second); err != nil {
		return true
	}

	return false
}

func (h *Handler) ComputerTasks(c echo.Context, successMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	currentPage := c.FormValue("page")
	pageSize := c.FormValue("pageSize")
	sortBy := c.FormValue("sortBy")
	sortOrder := c.FormValue("sortOrder")
	currentSortBy := c.FormValue("currentSortBy")

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(currentPage, pageSize, sortBy, sortOrder, currentSortBy, itemsPerPage)

	agent, err := h.Model.GetAgentNetworkAdaptersInfo(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	reports, err := h.Model.TaskReportsByPageInfo(agentId, commonInfo, p)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	p.NItems, err = h.Model.CountTaskReportsByPageInfo(agentId, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: an error occurred counting apps for agent: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	confirmDelete := c.QueryParam("delete") != ""

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	availableTasks, err := h.Model.GetAvailableTasksForAgent(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_available_tasks", err), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Tasks(c, p, agent, reports, availableTasks, confirmDelete, itemsPerPage, commonInfo, netbird, offline, successMessage, refreshTime), commonInfo))
}

func (h *Handler) RunTask(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	if c.FormValue("task") == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.edit.empty_task"), true))
	}

	taskSplitted := strings.Split(c.FormValue("task"), "-")
	taskID, err := strconv.Atoi(taskSplitted[len(taskSplitted)-1])
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.edit.invalid_task"), true))
	}

	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	t, err := h.Model.GetTasksById(taskID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.not_valid", err), true))
	}

	if t.Edges.Profile == nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.edit.no_profile", err), true))
	}

	if t.AgentType == task.AgentTypeLinux || t.AgentType == task.AgentTypeMacos {
		// prepare playbook
		pb, err := createAnsiblePlaybook(t)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%s : %v", i18n.T(c.Request().Context(), "tasks.could_not_create_ansible_playbook"), err), true))
		}

		// prepare request
		config := openuem_nats.ProfileConfig{
			ProfileID:     t.Edges.Profile.ID,
			AnsibleConfig: []*ansiblecfg.AnsiblePlaybook{pb},
		}

		data, err := yaml.Marshal(config)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.could_not_marshal_playbook", err), true))
		}

		// send request to agent
		if _, err = h.NATSConnection.Request("agent.ansible."+agentID, data, time.Duration(h.NATSTimeout)*time.Second); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.could_not_send_ansible_playbook_request", err), true))
		}
	}

	return h.ComputerTasks(c, i18n.T(c.Request().Context(), "tasks.task_was_run_successfully"))
}

func createAnsiblePlaybook(t *openuem_ent.Task) (*ansiblecfg.AnsiblePlaybook, error) {
	var err error

	pb := ansiblecfg.NewAnsiblePlaybook()
	pb.Name = t.Edges.Profile.Name

	switch t.Type {
	case task.TypeAddUnixLocalGroup:
		var gid int

		if t.LocalGroupID != "" {
			gid, err = strconv.Atoi(t.LocalGroupID)
			if err != nil {
				return nil, err
			}
		}

		addLocalGroup, err := ansiblecfg.AddLocalGroup(fmt.Sprintf("task_%d", t.ID), t.LocalGroupName, gid, t.LocalGroupSystem, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(addLocalGroup)

	case task.TypeAddUnixLocalUser:
		var expires float64
		var password_expire_account_disable int
		var password_expire_max int
		var password_expire_min int
		var password_expire_warn int
		var ssh_key_bits int
		var uid int
		var uid_max int
		var uid_min int

		if t.LocalUserExpires != "" {
			expires, err = strconv.ParseFloat(t.LocalUserExpires, 64)
			if err != nil {
				return nil, err
			}
		}

		if t.LocalUserPasswordExpireAccountDisable != "" {
			password_expire_account_disable, err = strconv.Atoi(t.LocalUserPasswordExpireAccountDisable)
			if err != nil {
				return nil, err
			}
		}

		if t.LocalUserPasswordExpireMax != "" {
			password_expire_max, err = strconv.Atoi(t.LocalUserPasswordExpireMax)
			if err != nil {
				return nil, err
			}
		}

		if t.LocalUserPasswordExpireMin != "" {
			password_expire_min, err = strconv.Atoi(t.LocalUserPasswordExpireMin)
			if err != nil {
				return nil, err
			}
		}

		if t.LocalUserPasswordExpireWarn != "" {
			password_expire_warn, err = strconv.Atoi(t.LocalUserPasswordExpireWarn)
			if err != nil {
				return nil, err
			}
		}

		if t.LocalUserSSHKeyBits != "" {
			ssh_key_bits, err = strconv.Atoi(t.LocalUserSSHKeyBits)
			if err != nil {
				return nil, err
			}
		}

		if t.LocalUserID != "" {
			uid, err = strconv.Atoi(t.LocalUserID)
			if err != nil {
				return nil, err
			}
		}

		if t.LocalUserIDMax != "" {
			uid_max, err = strconv.Atoi(t.LocalUserIDMax)
			if err != nil {
				return nil, err
			}
		}

		if t.LocalUserIDMin != "" {
			uid_min, err = strconv.Atoi(t.LocalUserIDMin)
			if err != nil {
				return nil, err
			}
		}

		addLinuxUser, err := ansiblecfg.AddLocalUser(fmt.Sprintf("task_%d", t.ID), t.LocalUserAppend, t.LocalUserDescription,
			t.LocalUserCreateHome, expires, t.LocalUserForce, t.LocalUserGenerateSSHKey, t.LocalUserGroup, t.LocalUserGroups,
			t.LocalUserHome, t.LocalUserUsername, t.LocalUserNonunique, t.LocalUserPassword, password_expire_account_disable, password_expire_max,
			password_expire_min, password_expire_warn, t.LocalUserPasswordLock, t.LocalUserShell, t.LocalUserSkeleton, ssh_key_bits,
			t.LocalUserSSHKeyComment, t.LocalUserSSHKeyFile, t.LocalUserSSHKeyPassphrase, t.LocalUserSSHKeyType,
			t.LocalUserSystem, t.LocalUserUmask, uid, uid_max, uid_min, t.AgentType.String(), t.IgnoreErrors)

		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(addLinuxUser)

	case task.TypeRemoveLocalUser:
		removeLinux, err := ansiblecfg.RemoveLocalUser(fmt.Sprintf("task_%d", t.ID), t.LocalUserForce, t.LocalUserUsername, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(removeLinux)

	case task.TypeRemoveUnixLocalGroup:
		removeLocalGroup, err := ansiblecfg.RemoveLocalGroup(fmt.Sprintf("task_%d", t.ID), t.LocalGroupName, t.LocalGroupForce, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(removeLocalGroup)

	case task.TypeUnixScript:
		executeScript, err := ansiblecfg.ExecuteScript(fmt.Sprintf("task_%d", t.ID), t.Script, t.ScriptExecutable, t.ScriptCreates, t.AgentType.String(), t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(executeScript)
	case task.TypeFlatpakInstall:
		install, err := ansiblecfg.InstallFlatpakPackage(fmt.Sprintf("task_%d", t.ID), t.PackageID, t.PackageLatest, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(install)
	case task.TypeFlatpakUninstall:
		uninstall, err := ansiblecfg.UninstallFlatpakPackage(fmt.Sprintf("task_%d", t.ID), t.PackageID, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(uninstall)
	case task.TypeBrewFormulaInstall:
		install, err := ansiblecfg.InstallHomeBrewFormula(fmt.Sprintf("task_%d", t.ID), t.PackageID, t.BrewInstallOptions, t.BrewUpdate, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(install)
	case task.TypeBrewFormulaUpgrade:
		upgrade, err := ansiblecfg.UpgradeHomeBrewFormula(fmt.Sprintf("task_%d", t.ID), t.PackageID, t.BrewUpdate, t.BrewUpgradeAll, t.BrewUpgradeOptions, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(upgrade)
	case task.TypeBrewFormulaUninstall:
		uninstall, err := ansiblecfg.UninstallHomeBrewFormula(fmt.Sprintf("task_%d", t.ID), t.PackageID, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(uninstall)
	case task.TypeBrewCaskInstall:
		install, err := ansiblecfg.InstallHomeBrewCask(fmt.Sprintf("task_%d", t.ID), t.PackageID, t.BrewInstallOptions, t.BrewUpdate, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(install)
	case task.TypeBrewCaskUpgrade:
		upgrade, err := ansiblecfg.UpgradeHomeBrewCask(fmt.Sprintf("task_%d", t.ID), t.PackageID, t.BrewGreedy, t.BrewUpdate, t.BrewUpgradeAll, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(upgrade)
	case task.TypeBrewCaskUninstall:
		uninstall, err := ansiblecfg.UninstallHomeBrewCask(fmt.Sprintf("task_%d", t.ID), t.PackageID, t.IgnoreErrors)
		if err != nil {
			return nil, err
		}
		pb.AddAnsibleTask(uninstall)
	}

	return pb, nil
}
