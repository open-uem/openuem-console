package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/doncicuto/openuem-console/internal/views/agents_views"
	"github.com/doncicuto/openuem-console/internal/views/computers_views"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListAgents(c echo.Context, successMessage, errMessage string) error {
	var err error
	var agents []*ent.Agent

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	// DEBUG - TEST set pageSize to 1
	// p.PageSize = 1

	// Get filters values
	f := filters.AgentFilter{}
	f.Hostname = c.FormValue("filterByHostname")

	enabledAgents := c.FormValue("filterByAgentsEnabled")
	if enabledAgents == "on" {
		f.EnabledAgents = true
	}

	disabledAgents := c.FormValue("filterByAgentsDisabled")
	if disabledAgents == "on" {
		f.DisabledAgents = true
	}

	windowsAgents := c.FormValue("filterByOSAgentWindows")
	if windowsAgents == "windows" {
		f.WindowsAgents = true
	}

	linuxAgents := c.FormValue("filterByOSAgentLinux")
	if linuxAgents == "linux" {
		f.LinuxAgents = true
	}

	macAgents := c.FormValue("filterByOSAgentMac")
	if macAgents == "mac" {
		f.MacAgents = true
	}

	contactFrom := c.FormValue("filterContactDateFrom")
	if contactFrom != "" {
		f.ContactFrom = contactFrom
	}
	contactTo := c.FormValue("filterContactDateTo")
	if contactTo != "" {
		f.ContactTo = contactTo
	}

	tags, err := h.Model.GetAppliedTags()
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	for _, tag := range tags {
		if c.FormValue(fmt.Sprintf("filterByTag%d", tag.ID)) != "" {
			f.Tags = append(f.Tags, tag.ID)
		}
	}

	tagId := c.FormValue("tagId")
	agentId := c.FormValue("agentId")
	if c.Request().Method == "POST" && tagId != "" && agentId != "" {
		err := h.Model.AddTagToAgent(agentId, tagId)
		if err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	if c.Request().Method == "DELETE" && tagId != "" && agentId != "" {
		err := h.Model.RemoveTagFromAgent(agentId, tagId)
		if err != nil {
			return renderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	agents, err = h.Model.GetAgentsByPage(p, f)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllAgents(f)
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return renderView(c, agents_views.AgentsIndex("| Agents", agents_views.Agents(c, p, f, agents, tags, successMessage, errMessage, refreshTime)))
}

func (h *Handler) AgentDelete(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return h.ListAgents(c, "", "an error ocurred getting uuid param")
	}

	agent, err := h.Model.GetAgentOSInfo(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error())
	}
	return renderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentsConfirmDelete(agent)))
}

func (h *Handler) AgentConfirmDelete(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return h.ListAgents(c, "", "an error ocurred getting uuid param")
	}

	err := h.Model.DeleteAgent(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error())
	}
	return h.ListAgents(c, "Agent was deleted successfully", "")
}

func (h *Handler) AgentEnable(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return fmt.Errorf("uuid cannot be empty")
	}

	if _, err := h.NATSConnection.Request("agent.enable."+agentId, nil, 10*time.Second); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.EnableAgent(agentId); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.has_been_enabled"), "")
}

func (h *Handler) AgentDisable(c echo.Context) error {
	agentId := c.Param("uuid")
	agent, err := h.Model.GetAgentOSInfo(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error())
	}
	return renderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentsConfirmDisable(agent)))
}

func (h *Handler) AgentForceRun(c echo.Context) error {
	agentId := c.Param("uuid")

	go func() {
		if _, err := h.NATSConnection.Request("agent.report."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second); err != nil {
			log.Printf("[ERROR]: %v", err)
		}
	}()

	return renderSuccess(c, partials.SuccessMessage(i18n.T(c.Request().Context(), "agents.force_run_success")))
}

func (h *Handler) AgentConfirmDisable(c echo.Context) error {
	agentId := c.Param("uuid")

	if _, err := h.NATSConnection.Request("agent.disable."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.DisableAgent(agentId); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.has_been_disabled"), "")
}

func (h *Handler) AgentStartVNC(c echo.Context) error {
	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentOSInfo(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error())
	}

	if _, err := h.NATSConnection.Request("agent.startvnc."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// TODO - Proxy port should not be hardcoded?
	return renderView(c, computers_views.VNC(agentId, agent.Hostname, "1443", h.Domain))

}

func (h *Handler) AgentStopVNC(c echo.Context) error {
	agentId := c.Param("uuid")

	if _, err := h.NATSConnection.Request("agent.stopvnc."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return renderView(c, computers_views.VNCConnect(agentId))

}
