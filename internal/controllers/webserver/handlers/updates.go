package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/open-uem/ent"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/computers_views"
	"github.com/open-uem/openuem-console/internal/views/deploy_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

const (
	// Listing updates can take a while: winget runs as SYSTEM and in every user session
	listPackageUpdatesTimeout = 10 * time.Minute
	// Agents answer as soon as an update run starts
	updatePackagesTimeout = 10 * time.Second
	// Maximum number of agents asked to update packages at the same time
	updatePackagesConcurrency = 16
	// Agents answer status requests without running the package manager
	updateStatusTimeout = 5 * time.Second
	agentPingTimeout    = 2 * time.Second

	// Subject of the update requests queued in AGENTS_STREAM (agent.update.>)
	queuedUpdatePackagesSubject = "agent.update.packages."
)

// DeployUpdate shows the page to update a package on several agents
func (h *Handler) DeployUpdate(c echo.Context) error {
	return h.DeployAction(c, deploy_views.ActionUpdate)
}

// DeployUpdateAll shows the agents where every package with an available update can be updated
func (h *Handler) DeployUpdateAll(c echo.Context) error {
	return h.SelectAgentsForDeployment(c, deploy_views.ActionUpdateAll)
}

// UpdateAllPackagesOnAgents queues an update run for every agent. The requests are kept in
// the agents' JetStream stream, so agents that are offline start the run when they connect.
func (h *Handler) UpdateAllPackagesOnAgents(c echo.Context, agentIDs []string, commonInfo *partials.CommonInfo) error {
	if h.JetStream == nil || h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	// Only agents of the current organization and site can be asked
	nicknames := map[string]string{}
	allowed := []string{}
	for _, agentID := range agentIDs {
		agent, err := h.Model.GetAgentById(agentID, commonInfo)
		if err != nil {
			continue
		}
		nicknames[agentID] = agent.Nickname
		allowed = append(allowed, agentID)
	}
	if len(allowed) == 0 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_selected_agents_to_deploy"), true))
	}

	result := deploy_views.UpdateAllResult{}
	for agentID, outcome := range h.queuePackageUpdates(allowed) {
		switch {
		case errors.Is(outcome.err, errUpdateNotQueued):
			result.Failures = append(result.Failures, deploy_views.UpdateRequestFailure{Computer: nicknames[agentID], Reason: i18n.T(c.Request().Context(), "updateall.not_queued")})
		case outcome.err != nil:
			result.Failures = append(result.Failures, deploy_views.UpdateRequestFailure{Computer: nicknames[agentID], Reason: outcome.err.Error()})
		case outcome.online:
			result.Started++
		default:
			result.Queued = append(result.Queued, nicknames[agentID])
		}
	}
	slices.Sort(result.Queued)

	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.UpdateAllRequested(c, result, commonInfo), commonInfo))
}

type queueOutcome struct {
	online bool
	err    error
}

// errUpdateNotQueued is returned when an offline agent can't receive queued requests, e.g.
// its version doesn't support it: AGENTS_STREAM only keeps messages agents are interested in
var errUpdateNotQueued = errors.New("the update request could not be queued")

// queuePackageUpdates publishes, concurrently, an update request for every agent in
// AGENTS_STREAM and checks which agents are online and start the run right away
func (h *Handler) queuePackageUpdates(agentIDs []string) map[string]queueOutcome {
	var mu sync.Mutex
	var wg sync.WaitGroup
	outcomes := map[string]queueOutcome{}
	limit := make(chan struct{}, updatePackagesConcurrency)

	for _, agentID := range agentIDs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()

			outcome := queueOutcome{}
			if outcome.err = h.queueAgentPackageUpdates(agentID); outcome.err == nil {
				outcome.online = h.agentIsOnline(agentID)
				if queued, known := h.queuedPackageUpdates(agentID); !outcome.online && known && !queued {
					outcome.err = errUpdateNotQueued
				}
			}

			mu.Lock()
			outcomes[agentID] = outcome
			mu.Unlock()
		}()
	}
	wg.Wait()

	return outcomes
}

func (h *Handler) queueAgentPackageUpdates(agentID string) error {
	data, err := json.Marshal(openuem_nats.UpdatePackagesRequest{AgentId: agentID, RequestedAt: time.Now()})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), updatePackagesTimeout)
	defer cancel()

	_, err = h.JetStream.Publish(ctx, queuedUpdatePackagesSubject+agentID, data)
	return err
}

func (h *Handler) agentIsOnline(agentID string) bool {
	_, err := h.NATSConnection.Request("agent.ping."+agentID, nil, agentPingTimeout)
	return err == nil
}

// queuedPackageUpdates reports if an update request is waiting in AGENTS_STREAM for an
// agent. Messages are removed from the stream once the agent has received them. known is
// false if the stream couldn't be read, e.g. because NATS doesn't allow it.
func (h *Handler) queuedPackageUpdates(agentID string) (queued bool, known bool) {
	if h.AgentStream == nil {
		return false, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), agentPingTimeout)
	defer cancel()

	_, err := h.AgentStream.GetLastMsgForSubject(ctx, queuedUpdatePackagesSubject+agentID)
	switch {
	case err == nil:
		return true, true
	case errors.Is(err, jetstream.ErrMsgNotFound):
		return false, true
	default:
		return false, false
	}
}

// DeployUpdateStatus shows the status of the software update runs of the agents
func (h *Handler) DeployUpdateStatus(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)
	if p.SortBy == "" {
		p.SortBy = "nickname"
		p.SortOrder = "asc"
	}

	f := filters.AgentFilter{}
	p.NItems, err = h.Model.CountAllAgents(f, true, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	agents, err := h.Model.GetAgentsByPage(p, f, true, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	rows := h.updateStatuses(agents)
	return RenderView(c, deploy_views.DeployIndex("| Deploy", deploy_views.UpdateStatus(c, p, rows, refreshTime, itemsPerPage, commonInfo), commonInfo))
}

// updateStatuses asks the agents, concurrently, for the status of their last update run
func (h *Handler) updateStatuses(agents []*ent.Agent) []deploy_views.UpdateStatusRow {
	var wg sync.WaitGroup
	rows := make([]deploy_views.UpdateStatusRow, len(agents))
	limit := make(chan struct{}, updatePackagesConcurrency)
	request, _ := json.Marshal(openuem_nats.ListPackageUpdatesRequest{StatusOnly: true})

	for i, agent := range agents {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()

			row := deploy_views.UpdateStatusRow{AgentID: agent.ID, Nickname: agent.Nickname, OS: agent.Os}

			msg, err := h.NATSConnection.Request("agent.listpackageupdates."+agent.ID, request, updateStatusTimeout)
			if err == nil {
				response := openuem_nats.AvailableUpdatesResponse{}
				if err := json.Unmarshal(msg.Data, &response); err == nil {
					row.Reachable = true
					row.Run = response.LastRun
				}
			} else {
				// Agents that don't support software updates are online but don't answer
				row.Online = h.agentIsOnline(agent.ID)
			}
			row.Queued, _ = h.queuedPackageUpdates(agent.ID)

			rows[i] = row
		}()
	}
	wg.Wait()

	return rows
}

// errAgentUpdateRefused is returned when the agent refuses to start an update run
type errAgentUpdateRefused struct{ reason string }

func (e errAgentUpdateRefused) Error() string { return e.reason }

func (h *Handler) requestAgentPackageUpdates(agentID string, packages []openuem_nats.AvailableUpdate) error {
	data, err := json.Marshal(openuem_nats.UpdatePackagesRequest{AgentId: agentID, Packages: packages})
	if err != nil {
		return err
	}

	msg, err := h.NATSConnection.Request("agent.updatepackages."+agentID, data, updatePackagesTimeout)
	if err != nil {
		return err
	}

	if len(msg.Data) > 0 {
		return errAgentUpdateRefused{reason: string(msg.Data)}
	}
	return nil
}

func updateRequestErrorMessage(c echo.Context, err error) string {
	var refused errAgentUpdateRefused
	switch {
	case errors.Is(err, nats.ErrNoResponders):
		return i18n.T(c.Request().Context(), "package_updates.agent_not_reachable")
	case errors.Is(err, nats.ErrTimeout):
		return i18n.T(c.Request().Context(), "package_updates.agent_timeout")
	case errors.As(err, &refused):
		return refused.reason
	default:
		return err.Error()
	}
}

// ComputerUpdates shows the page where the available updates of a computer are listed
func (h *Handler) ComputerUpdates(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentID, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}

	p := partials.NewPaginationAndSort(0)
	return RenderView(c, computers_views.InventoryIndex(" | Deploy", computers_views.ComputerUpdates(c, p, agent, commonInfo, settings.AccessToken != "", h.IsAgentOffline(c)), commonInfo))
}

// ComputerListUpdates asks the agent for its available updates and the result of its last update run
func (h *Handler) ComputerListUpdates(c echo.Context) error {
	return h.renderComputerAvailableUpdates(c, "")
}

// ComputerUpdatePackages asks the agent to update the selected packages, or every package if none is selected
func (h *Handler) ComputerUpdatePackages(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	// The agent must belong to the current organization and site
	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	packages := []openuem_nats.AvailableUpdate{}
	if c.FormValue("updateAll") != "true" {
		form, err := c.FormParams()
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		for _, value := range form["packages"] {
			p := openuem_nats.AvailableUpdate{}
			if err := json.Unmarshal([]byte(value), &p); err != nil {
				return RenderError(c, partials.ErrorMessage(err.Error(), false))
			}
			packages = append(packages, p)
		}

		if len(packages) == 0 {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "package_updates.no_packages_selected"), false))
		}
	}

	if err := h.requestAgentPackageUpdates(agentID, packages); err != nil {
		return RenderError(c, partials.ErrorMessage(updateRequestErrorMessage(c, err), false))
	}

	return h.renderComputerAvailableUpdates(c, i18n.T(c.Request().Context(), "package_updates.requested"))
}

func (h *Handler) renderComputerAvailableUpdates(c echo.Context, successMessage string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	// The agent must belong to the current organization and site
	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	msg, err := h.NATSConnection.Request("agent.listpackageupdates."+agentID, nil, listPackageUpdatesTimeout)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(updateRequestErrorMessage(c, err), false))
	}

	response := openuem_nats.AvailableUpdatesResponse{}
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return RenderError(c, partials.ErrorMessage(fmt.Sprintf("could not read the agent's response: %v", err), false))
	}

	return RenderView(c, computers_views.AvailableUpdates(agentID, &response, successMessage, commonInfo))
}
