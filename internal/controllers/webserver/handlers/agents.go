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
	"github.com/open-uem/openuem-console/internal/views"
	"github.com/open-uem/openuem-console/internal/views/agents_views"
	"github.com/open-uem/openuem-console/internal/views/computers_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/utils"
)

func (h *Handler) ListAgents(c echo.Context, successMessage, errMessage string, gotoPage1 bool) error {
	var err error
	var agents []*ent.Agent

	currentPage := c.FormValue("page")
	pageSize := c.FormValue("pageSize")
	sortBy := c.FormValue("sortBy")
	sortOrder := c.FormValue("sortOrder")
	currentSortBy := c.FormValue("currentSortBy")

	p := partials.NewPaginationAndSort()

	if gotoPage1 {
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

	if gotoPage1 {
		u, err := url.Parse(c.Request().Header.Get("Hx-Current-Url"))
		if err == nil {
			f.Hostname = u.Query().Get("filterByHostname")
		}
	} else {
		f.Hostname = c.FormValue("filterByHostname")
	}

	filteredAgentStatusOptions := []string{}
	for index := range agents_views.AgentStatus {
		if gotoPage1 {
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
		if gotoPage1 {
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

	if gotoPage1 {
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

	if gotoPage1 {
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

	if gotoPage1 {
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

	if gotoPage1 {
		currentUrl := c.Request().Header.Get("Hx-Current-Url")
		if currentUrl != "" {
			if u, err := url.Parse(currentUrl); err == nil {
				q := u.Query()
				q.Del("page")
				q.Add("page", "1")
				u.RawQuery = q.Encode()
				return RenderViewWithReplaceUrl(c, agents_views.AgentsIndex("| Agents", agents_views.Agents(c, p, f, h.SessionManager, l, agents, availableTags, appliedTags, availableOSes, successMessage, errMessage, refreshTime)), u)
			}
		}
	}

	return RenderView(c, agents_views.AgentsIndex("| Agents", agents_views.Agents(c, p, f, h.SessionManager, l, agents, availableTags, appliedTags, availableOSes, successMessage, errMessage, refreshTime)))
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
	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentsConfirmDelete(c, h.SessionManager, agent)))
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
	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentsConfirmDisable(c, h.SessionManager, agent)))
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

	return RenderConfirm(c, partials.ConfirmAdmitAgents(c.Request().Referer()))
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

	return RenderConfirm(c, partials.ConfirmEnableAgents(c.Request().Referer()))
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

	return RenderConfirm(c, partials.ConfirmDisableAgents(c.Request().Referer()))
}

func (h *Handler) AgentAdmit(c echo.Context) error {
	agentId := c.Param("uuid")
	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}
	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentConfirmAdmission(c, h.SessionManager, agent)))
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

func (h *Handler) AgentStartVNC(c echo.Context) error {
	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
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

		return RenderView(c, agents_views.AgentsIndex("| Agents", computers_views.VNC(agent, h.Domain, true, requestPIN, pin, h.SessionManager)))
	}

	return RenderView(c, agents_views.AgentsIndex("| Agents", computers_views.VNC(agent, h.Domain, false, false, "", h.SessionManager)))
}

func (h *Handler) AgentStopVNC(c echo.Context) error {
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

	return RenderView(c, agents_views.AgentsIndex("| Agents", computers_views.VNC(agent, h.Domain, false, false, "", h.SessionManager)))
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
