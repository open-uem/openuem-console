package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	openuem_nats "github.com/open-uem/nats"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views"
	"github.com/open-uem/openuem-console/internal/views/agents_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) ListAgents(c echo.Context, successMessage, errMessage string, comesFromDialog bool) error {
	var err error
	var agents []*ent.Agent

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

	filteredAgentStatusOptions := []string{}
	for index := range agents_views.AgentStatus {
		if comesFromDialog {
			u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
			if err == nil {
				value := u.Query().Get(fmt.Sprintf("filterByStatusAgent%d", index))
				if value != "" {
					filteredAgentStatusOptions = append(filteredAgentStatusOptions, value)
				}
			}
		} else {
			value := c.FormValue(fmt.Sprintf("filterByStatusAgent%d", index))
			if value != "" {
				filteredAgentStatusOptions = append(filteredAgentStatusOptions, value)
			}
		}
	}
	f.AgentStatusOptions = filteredAgentStatusOptions

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

	filteredIsRemote := []string{}
	for index := range []string{"Remote", "Local"} {
		value := c.FormValue(fmt.Sprintf("filterByIsRemote%d", index))
		if value != "" {
			filteredIsRemote = append(filteredIsRemote, value)
		}
	}
	f.IsRemote = filteredIsRemote

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			contactFrom := u.Query().Get("filterByContactDateFrom")
			if contactFrom != "" {
				f.ContactFrom = contactFrom
			}
			contactTo := u.Query().Get("filterByContactDateTo")
			if contactTo != "" {
				f.ContactTo = contactTo
			}
		}
	} else {
		contactFrom := c.FormValue("filterByContactDateFrom")
		if contactFrom != "" {
			f.ContactFrom = contactFrom
		}
		contactTo := c.FormValue("filterByContactDateTo")
		if contactTo != "" {
			f.ContactTo = contactTo
		}
	}

	availableTags, err := h.Model.GetAllTags()
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	appliedTags, err := h.Model.GetAppliedTags()
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			for _, tag := range appliedTags {
				if u.Query().Get(fmt.Sprintf("filterByTag%d", tag.ID)) != "" {
					f.Tags = append(f.Tags, tag.ID)
				}
			}
		}
	} else {
		for _, tag := range appliedTags {
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

	if comesFromDialog {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			nSelectedItems := u.Query().Get("filterBySelectedItems")
			f.SelectedItems, err = strconv.Atoi(nSelectedItems)
			if err != nil {
				f.SelectedItems = 0
			}
		}
	} else {
		nSelectedItems := c.FormValue("filterBySelectedItems")
		f.SelectedItems, err = strconv.Atoi(nSelectedItems)
		if err != nil {
			f.SelectedItems = 0
		}
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

	agents, err = h.Model.GetAgentsByPage(p, f, false)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	p.NItems, err = h.Model.CountAllAgents(f, false)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	l := views.GetTranslatorForDates(c)

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if comesFromDialog {
		currentUrl := c.Request().Header.Get("Hx-Current-Url")
		if currentUrl != "" {
			if u, err := url.Parse(currentUrl); err == nil {
				q := u.Query()
				q.Del("page")
				q.Add("page", "1")
				u.RawQuery = q.Encode()
				return RenderViewWithReplaceUrl(c, agents_views.AgentsIndex("| Agents", agents_views.Agents(c, p, f, h.SessionManager, l, h.Version, latestServerRelease.Version, agents, availableTags, appliedTags, availableOSes, successMessage, errMessage, refreshTime)), u)
			}
		}
	}

	return RenderView(c, agents_views.AgentsIndex("| Agents", agents_views.Agents(c, p, f, h.SessionManager, l, h.Version, latestServerRelease.Version, agents, availableTags, appliedTags, availableOSes, successMessage, errMessage, refreshTime)))
}

func (h *Handler) AgentDelete(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return h.ListAgents(c, "", "an error ocurred getting uuid param", false)
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentsConfirmDelete(c, h.SessionManager, h.Version, latestServerRelease.Version, agent)))
}

func (h *Handler) AgentConfirmDelete(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return h.ListAgents(c, "", "an error ocurred getting uuid param", false)
	}

	err := h.Model.DeleteAgent(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.deleted"), "", true)
}

func (h *Handler) AgentEnable(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return fmt.Errorf("uuid cannot be empty")
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := h.JetStream.Publish(ctx, "agent.enable."+agentId, nil); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.EnableAgent(agentId); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.has_been_enabled"), "", true)
}

func (h *Handler) AgentDisable(c echo.Context) error {
	agentId := c.Param("uuid")
	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentsConfirmDisable(c, h.SessionManager, h.Version, latestServerRelease.Version, agent)))
}

func (h *Handler) AgentsAdmit(c echo.Context) error {
	errorsFound := false

	if c.Request().Method == "POST" {
		agents := c.FormValue("agents")

		for _, agentId := range strings.Split(agents, ",") {

			agent, err := h.Model.GetAgentById(agentId)
			if err != nil {
				log.Println("[ERROR]: ", i18n.T(c.Request().Context(), "agents.not_found"))
				errorsFound = true
				continue
			}

			if agent.AgentStatus == "WaitingForAdmission" {

				if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
					log.Println("[ERROR]: ", i18n.T(c.Request().Context(), "nats.not_connected"))
					errorsFound = true
					continue
				}

				data, err := json.Marshal(openuem_nats.CertificateRequest{
					AgentId:      agentId,
					DNSName:      agent.Hostname + "." + h.Domain,
					Organization: h.OrgName,
					Province:     h.OrgProvince,
					Locality:     h.OrgLocality,
					Address:      h.OrgAddress,
					Country:      h.Country,
					YearsValid:   2,
				})
				if err != nil {
					log.Println("[ERROR]: ", err.Error())
					errorsFound = true
					continue
				}

				if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
					log.Println("[ERROR]: ", i18n.T(c.Request().Context(), "nats.not_connected"))
					errorsFound = true
					continue
				}

				if _, err := h.NATSConnection.Request("certificates.agent."+agentId, data, time.Duration(h.NATSTimeout)*time.Second); err != nil {
					log.Println("[ERROR]: ", i18n.T(c.Request().Context(), "nats.no_responder"))
					errorsFound = true
					continue
				}

				if err := h.Model.EnableAgent(agentId); err != nil {
					log.Println("[ERROR]: ", err.Error())
					errorsFound = true
					continue
				}

				if settings, err := h.Model.GetGeneralSettings(); err != nil {
					log.Println("[ERROR]: ", err.Error())
					errorsFound = true
					continue
				} else {
					if settings.Edges.Tag != nil {
						if err := h.Model.AddTagToAgent(agentId, strconv.Itoa(settings.Edges.Tag.ID)); err != nil {
							log.Println("[ERROR]: ", err.Error())
							errorsFound = true
							continue
						}
					}
				}

			} else {
				log.Printf("[ERROR]: agent %s is not in a valid state\n", agentId)
				errorsFound = true
				continue
			}
		}

		if errorsFound {
			return h.ListAgents(c, "", i18n.T(c.Request().Context(), "agents.some_could_not_be_admitted"), true)
		}
		return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.have_been_admitted"), "", true)
	}

	return RenderConfirm(c, partials.ConfirmAdmitAgents(c))
}

func (h *Handler) AgentsEnable(c echo.Context) error {
	errorsFound := false

	if c.Request().Method == "POST" {
		agents := c.FormValue("agents")

		for _, agentId := range strings.Split(agents, ",") {
			agent, err := h.Model.GetAgentById(agentId)
			if err != nil {
				log.Println("[ERROR]: ", err.Error())
				errorsFound = true
				continue
			}

			if agent.AgentStatus == "Disabled" {
				if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
					log.Println("[ERROR]: ", i18n.T(c.Request().Context(), "nats.not_connected"))
					errorsFound = true
					continue
				}

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if _, err := h.JetStream.Publish(ctx, "agent.enable."+agentId, nil); err != nil {
					log.Println("[ERROR]: ", err.Error())
					errorsFound = true
					continue
				}

				if err := h.Model.EnableAgent(agentId); err != nil {
					log.Println("[ERROR]: ", err.Error())
					errorsFound = true
					continue
				}
			} else {
				log.Printf("[ERROR]: agent %s is not in a valid state\n", agentId)
				errorsFound = true
				continue
			}
		}
		if errorsFound {
			return h.ListAgents(c, "", i18n.T(c.Request().Context(), "agents.some_could_not_be_enabled"), true)
		}
		return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.have_been_enabled"), "", true)
	}

	return RenderConfirm(c, partials.ConfirmEnableAgents(c))
}

func (h *Handler) AgentsDisable(c echo.Context) error {
	errorsFound := false

	if c.Request().Method == "POST" {

		agents := c.FormValue("agents")

		for _, agentId := range strings.Split(agents, ",") {
			agent, err := h.Model.GetAgentById(agentId)
			if err != nil {
				log.Println("[ERROR]: ", err.Error())
				errorsFound = true
				continue
			}

			if agent.AgentStatus == "Enabled" {
				if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
				}

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if _, err := h.JetStream.Publish(ctx, "agent.disable."+agentId, nil); err != nil {
					return RenderError(c, partials.ErrorMessage(err.Error(), false))
				}

				if err := h.Model.DisableAgent(agentId); err != nil {
					return RenderError(c, partials.ErrorMessage(err.Error(), false))
				}
			} else {
				log.Printf("[ERROR]: agent %s is not in a valid state\n", agentId)
				errorsFound = true
				continue
			}
		}
		if errorsFound {
			return h.ListAgents(c, "", i18n.T(c.Request().Context(), "agents.some_could_not_be_disabled"), true)
		}
		return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.have_been_disabled"), "", true)

	}

	return RenderConfirm(c, partials.ConfirmDisableAgents(c))
}

func (h *Handler) AgentAdmit(c echo.Context) error {
	agentId := c.Param("uuid")
	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentConfirmAdmission(c, h.SessionManager, h.Version, latestServerRelease.Version, agent)))
}

func (h *Handler) AgentForceRun(c echo.Context) error {
	agentId := c.Param("uuid")

	go func() {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			log.Printf("[ERROR]: %s", i18n.T(c.Request().Context(), "nats.not_connected"))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := h.JetStream.Publish(ctx, "agent.report."+agentId, nil); err != nil {
			log.Printf("[ERROR]: %v", err)
		}
	}()

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.force_run_success"), "", false)
}

func (h *Handler) AgentConfirmDisable(c echo.Context) error {
	agentId := c.Param("uuid")

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := h.JetStream.Publish(ctx, "agent.disable."+agentId, nil); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.DisableAgent(agentId); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.has_been_disabled"), "", true)
}

func (h *Handler) AgentConfirmAdmission(c echo.Context, regenerate bool) error {
	agentId := c.Param("uuid")
	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_found"), false))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	data, err := json.Marshal(openuem_nats.CertificateRequest{
		AgentId:      agentId,
		DNSName:      agent.Hostname + "." + h.Domain,
		Organization: h.OrgName,
		Province:     h.OrgProvince,
		Locality:     h.OrgLocality,
		Address:      h.OrgAddress,
		Country:      h.Country,
		YearsValid:   2,
	})
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
	}

	if _, err := h.NATSConnection.Request("certificates.agent."+agentId, data, time.Duration(h.NATSTimeout)*time.Second); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responder"), false))
	}

	if err := h.Model.EnableAgent(agentId); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if regenerate {
		return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.certs_regenerated"), "", false)
	}
	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.has_been_admitted"), "", true)
}

func (h *Handler) AgentForceRestart(c echo.Context) error {
	agentId := c.Param("uuid")

	if c.Request().Method == "POST" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		if _, err := h.NATSConnection.Request("agent.restart."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responder"), false))
		}
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.has_been_restarted"), "", false)
}

func (h *Handler) AgentEnableDebug(c echo.Context) error {
	agentId := c.Param("uuid")

	if c.Request().Method == "POST" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		msg, err := h.NATSConnection.Request("agent.enabledebug."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responder"), false))
		}

		if string(msg.Data) == "enabled" {
			if err := h.Model.EnableDebugAgent(agentId); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_save_debug_mode"), false))
			}
		}
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.debug_has_been_enabled"), "", false)
}

func (h *Handler) AgentDisableDebug(c echo.Context) error {
	agentId := c.Param("uuid")

	if c.Request().Method == "POST" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		msg, err := h.NATSConnection.Request("agent.disabledebug."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responder"), false))
		}

		if string(msg.Data) == "disabled" {
			if err := h.Model.DisableDebugAgent(agentId); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_save_debug_mode"), false))
			}
		}
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.debug_has_been_disabled"), "", false)
}
