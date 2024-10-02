package handlers

import (
	"log"
	"slices"
	"strconv"

	"github.com/doncicuto/openuem-console/internal/views/desktops_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
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

	agentId := c.Param("uuid")

	if agentId == "" {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error("an error ocurred getting uuid param", "Desktop", "/desktops")))
	}

	currentPage := 1
	if currentPage, err = strconv.Atoi(c.QueryParam("page")); err != nil {
		currentPage = 1
	}

	pageSize := 5
	if pageSize, err = strconv.Atoi(c.QueryParam("pageSize")); err != nil {
		pageSize = 5
	}
	// TODO set a MAX value for pageSize
	pageSize = min(pageSize, 25)

	sortBy := c.QueryParam("sortBy")
	if !slices.Contains([]string{"hostname", "os", "version", "last_contact"}, sortBy) {
		sortBy = ""
	}

	sortOrder := c.QueryParam("sortOrder")
	if !slices.Contains([]string{"asc", "desc"}, sortOrder) {
		sortOrder = "asc"
	}

	a, err := h.Model.GetAgentById(agentId)
	if err != nil {
		log.Fatalf("an error ocurred querying agent: %v", err)
	}

	apps, err := h.Model.GetAgentAppsByPage(agentId, currentPage, pageSize)
	if err != nil {
		log.Fatalf("an error ocurred querying apps for agent: %v", err)
	}

	nApps, err := h.Model.CountAgentApps(agentId)
	if err != nil {
		log.Fatalf("an error ocurred querying apps for agent: %v", err)
	}

	confirmDelete := c.QueryParam("delete") != ""
	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.Apps(a, apps, c, currentPage, pageSize, nApps, sortBy, sortOrder, confirmDelete)))
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

	currentPage := 1
	if currentPage, err = strconv.Atoi(c.QueryParam("page")); err != nil {
		currentPage = 1
	}

	pageSize := 5
	if pageSize, err = strconv.Atoi(c.QueryParam("pageSize")); err != nil {
		pageSize = 5
	}
	// TODO set a MAX value for pageSize
	pageSize = min(pageSize, 25)

	sortBy := c.QueryParam("sortBy")
	if !slices.Contains([]string{"hostname", "os", "version", "username", "manufacturer", "model"}, sortBy) {
		sortBy = "hostname"
	}

	sortOrder := c.QueryParam("sortOrder")
	if !slices.Contains([]string{"asc", "desc"}, sortOrder) {
		sortOrder = "asc"
	}

	totalItems := 0
	totalItems, err = h.Model.CountAllAgents()
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	desktops, err := h.Model.GetDesktopsByPage(currentPage, pageSize, sortBy, sortOrder)
	if err != nil {
		return renderView(c, desktops_views.InventoryIndex(" | Inventory", partials.Error(err.Error(), "Desktops", "/desktops")))
	}

	return renderView(c, desktops_views.InventoryIndex(" | Inventory", desktops_views.Desktops(desktops, c, currentPage, pageSize, totalItems, sortBy, sortOrder)))
}
