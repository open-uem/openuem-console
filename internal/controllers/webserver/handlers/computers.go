package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"

	models "github.com/doncicuto/openuem-console/internal/models/winget"
	"github.com/doncicuto/openuem-console/internal/views/agents_views"
	"github.com/doncicuto/openuem-console/internal/views/computers_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_nats"
	"github.com/gomarkdown/markdown"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/mdlayher/wol"
	"github.com/microcosm-cc/bluemonday"
)

func (h *Handler) Computer(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentComputerInfo(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Computer(agent, confirmDelete)))
}

func (h *Handler) OperatingSystem(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentOSInfo(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.OperatingSystem(agent, confirmDelete)))
}

func (h *Handler) NetworkAdapters(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentNetworkAdaptersInfo(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.NetworkAdapters(agent, confirmDelete)))
}

func (h *Handler) Printers(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentPrintersInfo(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Printers(agent, confirmDelete)))
}

func (h *Handler) LogicalDisks(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentLogicalDisksInfo(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.LogicalDisks(agent, confirmDelete)))
}

func (h *Handler) Shares(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentSharesInfo(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Shares(agent, confirmDelete)))
}

func (h *Handler) Monitors(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentMonitorsInfo(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Monitors(agent, confirmDelete)))
}

func (h *Handler) Apps(c echo.Context) error {
	var err error
	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	a, err := h.Model.GetAgentById(agentId)
	if err != nil {
		log.Fatalf("an error ocurred querying agent: %v", err)
	}

	apps, err := h.Model.GetAgentAppsByPage(agentId, p)
	if err != nil {
		log.Fatalf("an error ocurred querying apps for agent: %v", err)
	}

	p.NItems, err = h.Model.CountAgentApps(agentId)
	if err != nil {
		log.Fatalf("an error ocurred querying apps for agent: %v", err)
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Apps(c, p, a, apps, confirmDelete)))
}

func (h *Handler) RemoteAssistance(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentMonitorsInfo(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.RemoteAssistance(agent, confirmDelete)))
}

func (h *Handler) Computers(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "hostname"
		p.SortOrder = "desc"
	}

	computers, err := h.Model.GetComputersByPage(p)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	p.NItems, err = h.Model.CountAllAgents(agents_views.AgentFilter{})
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Computers(c, p, computers)))
}

func (h *Handler) ComputerDeploy(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderError(c, partials.ErrorMessage("an error ocurred getting uuid param", false))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	confirmDelete := c.QueryParam("delete") != ""

	deployments, err := h.Model.GetDeploymentsForAgent(agentId, p)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountDeploymentsForAgent(agentId)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		return renderView(c, computers_views.DeploymentsTable(c, p, agentId, deployments))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		return err
	}

	return renderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.ComputerDeploy(c, p, agent, deployments, confirmDelete, refreshTime)))
}

func (h *Handler) ComputerDeploySearchPackagesInstall(c echo.Context) error {
	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	agentId := c.Param("uuid")
	if agentId == "" {
		return renderError(c, partials.ErrorMessage("agent id cannot be empty", false))
	}

	search := c.FormValue("search")

	if search == "" {
		return renderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "install.search_empty_error"), true))
	}

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	packages, err := models.SearchPackages(search, p)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p.NItems, err = models.CountPackages(search)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderView(c, computers_views.SearchPacketResult(c, agentId, packages, p))
}

func (h *Handler) ComputerDeployInstall(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return renderError(c, partials.ErrorMessage("agent id cannot be empty", false))
	}

	packageId := c.FormValue("packageId")
	packageName := c.FormValue("packageName")

	if packageId == "" || packageName == "" {
		return renderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.deploy_empty_values"), true))
	}

	alreadyInstalled, err := h.Model.DeploymentAlreadyInstalled(agentId, packageId)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if alreadyInstalled {
		return renderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.already_deployed"), true))
	}

	action := openuem_nats.DeployAction{}
	action.AgentId = agentId
	action.PackageId = packageId
	action.PackageName = packageName
	action.Repository = "winget"
	action.Action = "install"

	data, err := json.Marshal(action)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	err = h.NATSConnection.Publish("agent.installpackage."+agentId, data)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.deploy_success")))
}

func (h *Handler) ComputerDeployUpdate(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return renderError(c, partials.ErrorMessage("agent id cannot be empty", false))
	}

	packageId := c.FormValue("packageId")
	packageName := c.FormValue("packageName")

	if packageId == "" || packageName == "" {
		return renderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.deploy_empty_values"), true))
	}

	action := openuem_nats.DeployAction{}
	action.AgentId = agentId
	action.PackageId = packageId
	action.PackageName = packageName
	action.Repository = "winget"
	action.Action = "update"

	data, err := json.Marshal(action)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	err = h.NATSConnection.Publish("agent.updatepackage."+agentId, data)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.update_success")))
}

func (h *Handler) ComputerDeployUninstall(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return renderError(c, partials.ErrorMessage("agent id cannot be empty", false))
	}

	packageId := c.FormValue("packageId")
	packageName := c.FormValue("packageName")

	if packageId == "" || packageName == "" {
		return renderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.deploy_empty_values"), true))
	}

	action := openuem_nats.DeployAction{}
	action.AgentId = agentId
	action.PackageId = packageId
	action.PackageName = packageName
	action.Repository = "winget"
	action.Action = "uninstall"

	data, err := json.Marshal(action)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	err = h.NATSConnection.Publish("agent.uninstallpackage."+agentId, data)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.uninstall_success")))
}

func (h *Handler) WakeOnLan(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return renderError(c, partials.ErrorMessage("agent id cannot be empty", false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	if c.Request().Method == "GET" {
		confirmDelete := c.QueryParam("delete") != ""
		return renderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.WakeOnLan(agent, confirmDelete)))
	}

	mac := c.FormValue("MACAddress")
	hwAddress, err := net.ParseMAC(mac)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	ip := c.FormValue("IPAddress")

	wolClient, err := wol.NewClient()
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	err = wolClient.Wake(ip+":0", hwAddress)
	if err != nil {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Computers", "/computers")))
	}

	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.wol_success")))
}

func (h *Handler) ComputerMetadata(c echo.Context) error {
	var data []*openuem_ent.Metadata
	successMessage := ""

	agentId := c.Param("uuid")

	if agentId == "" {
		return renderError(c, partials.ErrorMessage("an error ocurred getting uuid param", false))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	if p.SortBy == "" {
		p.SortBy = "name"
		p.SortOrder = "asc"
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	confirmDelete := c.QueryParam("delete") != ""

	data, err = h.Model.GetMetadataForAgent(agentId, p)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	orgMetadata, err := h.Model.GetAllOrgMetadata()
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllOrgMetadata()
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		orgMetadataId := c.FormValue("orgMetadataId")
		name := c.FormValue("name")
		value := c.FormValue("value")

		id, err := strconv.Atoi(orgMetadataId)
		if err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}

		if orgMetadataId != "" && name != "" && value != "" {
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
				return renderError(c, partials.ErrorMessage(fmt.Sprintf("%s is not an accepted metadata", name), false))
			}

			if err := h.Model.SaveMetadata(agentId, id, value); err != nil {
				return renderError(c, partials.ErrorMessage(err.Error(), false))
			}

			data, err = h.Model.GetMetadataForAgent(agentId, p)
			if err != nil {
				return renderError(c, partials.ErrorMessage(err.Error(), false))
			}

			successMessage = i18n.T(c.Request().Context(), "agents.metadata_save_success")
		}
	}

	return renderView(c, computers_views.InventoryIndex(" | Deploy SW", computers_views.ComputerMetadata(c, p, agent, data, orgMetadata, confirmDelete, successMessage)))
}

func (h *Handler) Notes(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Computer", "/computers")))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if c.Request().Method == "POST" {
		notes := c.FormValue("markdown")
		if err := h.Model.SaveNotes(agentId, notes); err != nil {
			return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "notes.error", err.Error())))
		}
		return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "notes.updated")))
	}

	maybeUnsafeHTML := markdown.ToHTML([]byte(agent.Notes), nil, nil)
	renderedMarkdown := string(bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML))

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Notes(c, agent, agent.Notes, renderedMarkdown, confirmDelete)))
}
