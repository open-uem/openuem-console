package handlers

import (
	"encoding/json"
	"log"

	models "github.com/doncicuto/openuem-console/internal/models/winget"
	"github.com/doncicuto/openuem-console/internal/views/agents_views"
	"github.com/doncicuto/openuem-console/internal/views/desktops_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_nats"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Computer(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	agent, err := h.Model.GetAgentComputerInfo(agentId)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.Computer(agent, confirmDelete)))
}

func (h *Handler) OperatingSystem(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	agent, err := h.Model.GetAgentOSInfo(agentId)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.OperatingSystem(agent, confirmDelete)))
}

func (h *Handler) NetworkAdapters(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	agent, err := h.Model.GetAgentNetworkAdaptersInfo(agentId)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.NetworkAdapters(agent, confirmDelete)))
}

func (h *Handler) Printers(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	agent, err := h.Model.GetAgentPrintersInfo(agentId)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.Printers(agent, confirmDelete)))
}

func (h *Handler) LogicalDisks(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	agent, err := h.Model.GetAgentLogicalDisksInfo(agentId)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.LogicalDisks(agent, confirmDelete)))
}

func (h *Handler) Shares(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	agent, err := h.Model.GetAgentSharesInfo(agentId)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.Shares(agent, confirmDelete)))
}

func (h *Handler) Monitors(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	agent, err := h.Model.GetAgentMonitorsInfo(agentId)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.Monitors(agent, confirmDelete)))
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
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
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
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.Apps(c, p, a, apps, confirmDelete)))
}

func (h *Handler) RemoteAssistance(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	agent, err := h.Model.GetAgentMonitorsInfo(agentId)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.RemoteAssistance(agent, confirmDelete)))
}

func (h *Handler) Desktops(c echo.Context) error {
	var err error

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "hostname"
		p.SortOrder = "desc"
	}

	desktops, err := h.Model.GetDesktopsByPage(p)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	p.NItems, err = h.Model.CountAllAgents(agents_views.AgentFilter{})
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.Desktops(c, p, desktops)))
}

func (h *Handler) DesktopDeploy(c echo.Context) error {
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
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
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
		return renderView(c, desktops_views.DeploymentsTable(c, p, agentId, deployments))
	}

	return renderView(c, desktops_views.InventoryIndex(" | Deploy SW", desktops_views.DesktopDeploy(c, p, agent, deployments, confirmDelete)))
}

func (h *Handler) DesktopDeploySearchPackagesInstall(c echo.Context) error {
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

	return renderView(c, desktops_views.SearchPacketResult(c, agentId, packages, p))
}

func (h *Handler) DesktopDeployInstall(c echo.Context) error {
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

	err = h.MessageServer.Connection.Publish("agent.installpackage."+agentId, data)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.deploy_success")))
}

func (h *Handler) DesktopDeployUpdate(c echo.Context) error {
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

	err = h.MessageServer.Connection.Publish("agent.updatepackage."+agentId, data)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.update_success")))
}

func (h *Handler) DesktopDeployUninstall(c echo.Context) error {
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

	err = h.MessageServer.Connection.Publish("agent.uninstallpackage."+agentId, data)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.uninstall_success")))
}
