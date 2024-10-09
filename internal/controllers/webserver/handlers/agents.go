package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/doncicuto/openuem-console/internal/views/agents_views"
	"github.com/doncicuto/openuem-console/internal/views/desktops_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListAgents(c echo.Context, successMessage, errMessage string) error {
	var err error
	var agents []*ent.Agent

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	p.NItems, err = h.Model.CountAllAgents()
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	agents, err = h.Model.GetAgentsByPage(p)
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	return renderView(c, agents_views.AgentsIndex("| Agents", agents_views.Agents(c, agents, p, successMessage, errMessage)))
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

	if _, err := h.MessageServer.Connection.Request("agent.enable."+agentId, nil, 10*time.Second); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.EnableAgent(agentId); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListAgents(c, "Agent has been enabled", "")
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

	// TODO - Timeout should not be hardcoded
	go func() {
		if _, err := h.MessageServer.Connection.Request("agent.report."+agentId, nil, 10*time.Second); err != nil {
			log.Printf("[ERROR]: %v", err)
		}
	}()

	return h.ListAgents(c, "Agent will run an send new information, check it again in a few minutes", "")
}

func (h *Handler) AgentConfirmDisable(c echo.Context) error {
	agentId := c.Param("uuid")

	// TODO - Timeout should not be hardcoded
	if _, err := h.MessageServer.Connection.Request("agent.disable."+agentId, nil, 10*time.Second); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.DisableAgent(agentId); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListAgents(c, "Agent has been disabled", "")
}

func (h *Handler) AgentStartVNC(c echo.Context) error {
	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentOSInfo(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error())
	}

	// TODO - Timeout should not be hardcoded
	if _, err := h.MessageServer.Connection.Request("agent.startvnc."+agentId, nil, 120*time.Second); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// TODO - Domain should not be hardcoded
	return renderView(c, desktops_views.VNC(agentId, agent.Hostname, "1443", ".openuem.eu"))

}

func (h *Handler) AgentStopVNC(c echo.Context) error {
	agentId := c.Param("uuid")

	// TODO - Timeout should not be hardcoded
	if _, err := h.MessageServer.Connection.Request("agent.stopvnc."+agentId, nil, 120*time.Second); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// TODO - Domain should not be hardcoded
	return renderView(c, desktops_views.VNCConnect(agentId))

}
