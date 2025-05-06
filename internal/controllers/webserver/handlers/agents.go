package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/agents_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/utils"
)

func (h *Handler) ListAgents(c echo.Context, successMessage, errMessage string, comesFromDialog bool) error {
	var err error
	var agents []*ent.Agent

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

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
					if value == "No Contact" {
						f.NoContact = true
					}
					filteredAgentStatusOptions = append(filteredAgentStatusOptions, value)
				}
			}
		} else {
			value := c.FormValue(fmt.Sprintf("filterByStatusAgent%d", index))
			if value != "" {
				if value == "No Contact" {
					f.NoContact = true
				}
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

	sftpDisabled, err := h.Model.GetDefaultSFTPDisabled()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_sftp_general_setting"), true))
	}

	if comesFromDialog {
		currentUrl := c.Request().Header.Get("Hx-Current-Url")
		if currentUrl != "" {
			if u, err := url.Parse(currentUrl); err == nil {
				q := u.Query()
				q.Del("page")
				q.Add("page", "1")
				u.RawQuery = q.Encode()
				return RenderViewWithReplaceUrl(c, agents_views.AgentsIndex("| Agents", agents_views.Agents(c, p, f, agents, availableTags, appliedTags, availableOSes, sftpDisabled, successMessage, errMessage, refreshTime, commonInfo), commonInfo), u)
			}
		}
	}

	return RenderView(c, agents_views.AgentsIndex("| Agents", agents_views.Agents(c, p, f, agents, availableTags, appliedTags, availableOSes, sftpDisabled, successMessage, errMessage, refreshTime, commonInfo), commonInfo))
}

func (h *Handler) AgentDelete(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), true))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}

	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentsConfirmDelete(c, agent, commonInfo), commonInfo))
}

func (h *Handler) AgentConfirmDelete(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), true))
	}

	deleteAction := c.FormValue("agent-delete-action")

	if deleteAction == "delete-and-uninstall" || deleteAction == "keep-and-uninstall" {
		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := h.JetStream.Publish(ctx, "agent.uninstall."+agentId, nil); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_request_to_uninstall"), true))
		}
	}

	if deleteAction == "delete-and-uninstall" || deleteAction == "delete-and-keep" {
		err := h.Model.DeleteAgent(agentId)
		if err != nil {
			return h.ListAgents(c, "", err.Error(), false)
		}
	}

	return h.ListAgents(c, i18n.T(c.Request().Context(), "agents.deleted"), "", true)
}

func (h *Handler) AgentEnable(c echo.Context) error {
	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), true))
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
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}

	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentsConfirmDisable(c, agent, commonInfo), commonInfo))
}

func (h *Handler) AgentsAdmit(c echo.Context) error {
	errorsFound := false

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if c.Request().Method == "POST" {
		agents := c.FormValue("agents")

		for agentId := range strings.SplitSeq(agents, ",") {

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

				if err := h.NATSConnection.Publish("certificates.agent."+agentId, data); err != nil {
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

	return RenderConfirm(c, partials.ConfirmAdmitAgents(c, commonInfo))
}

func (h *Handler) AgentsEnable(c echo.Context) error {
	errorsFound := false

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if c.Request().Method == "POST" {
		agents := c.FormValue("agents")

		for agentId := range strings.SplitSeq(agents, ",") {
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

	return RenderConfirm(c, partials.ConfirmEnableAgents(c, commonInfo))
}

func (h *Handler) AgentsDisable(c echo.Context) error {
	errorsFound := false

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if c.Request().Method == "POST" {

		agents := c.FormValue("agents")

		for agentId := range strings.SplitSeq(agents, ",") {
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

	return RenderConfirm(c, partials.ConfirmDisableAgents(c, commonInfo))
}

func (h *Handler) AgentAdmit(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")
	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return h.ListAgents(c, "", err.Error(), false)
	}

	return RenderView(c, agents_views.AgentsIndex(" | Agents", agents_views.AgentConfirmAdmission(c, agent, commonInfo), commonInfo))
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
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
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

	if err := h.NATSConnection.Publish("certificates.agent."+agentId, data); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.no_responder"), false))
	}

	if err := h.Model.EnableAgent(agentId); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	sftpServiceDisabled, err := h.Model.GetDefaultSFTPDisabled()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	remoteAssistanceDisabled, err := h.Model.GetDefaultRemoteAssistanceDisabled()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := h.Model.Client.Agent.UpdateOneID(agentId).SetSftpService(!sftpServiceDisabled).SetRemoteAssistance(!remoteAssistanceDisabled).Exec(context.Background()); err != nil {
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

func (h *Handler) AgentLogs(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	category := c.FormValue("log-category")

	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), true))
	}

	a, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent", err.Error()), true))
	}

	logFile := ""
	if a.Os == "windows" {
		logFile = "C:\\Program Files\\OpenUEM Agent\\logs\\openuem-log.txt"
	} else {
		logFile = "/var/log/openuem-agent/openuem-agent.log"
	}

	// Get agents log using SFTP
	data, err := h.GetAgentLogFile(a, logFile)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_log", err.Error()), true))
	}

	agentLog := parseLogFile(data, category)

	if a.Os == "windows" {
		logFile = "C:\\Program Files\\OpenUEM Agent\\logs\\openuem-agent-updater.txt"
	} else {
		logFile = "/var/log/openuem-agent/openuem-updater.log"
	}

	// Get updaters log using SFTP
	data, err = h.GetAgentLogFile(a, logFile)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_log", err.Error()), true))
	}

	updaterLog := parseLogFile(data, category)

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	return RenderView(c, agents_views.AgentsIndex("| Agents", agents_views.AgentsLog(c, a, agentLog, updaterLog, category, "", "", refreshTime, commonInfo), commonInfo))
}

func (h *Handler) GetAgentLogFile(agent *ent.Agent, path string) (string, error) {
	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return "", err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return "", err
	}
	defer client.Close()
	defer sshConn.Close()

	dstFile, err := os.CreateTemp(h.DownloadDir, "*.log")
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	srcFile, err := client.OpenFile(path, (os.O_RDONLY))
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return "", err
	}

	// Read file
	l, err := os.ReadFile(dstFile.Name())
	if err != nil {
		return "", err
	}

	return string(l), nil
}

func parseLogFile(data, category string) []agents_views.LogEntry {
	logEntries := []agents_views.LogEntry{}
	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.Contains(line, ">>>>>>>") || strings.Contains(line, "<<<<<<<") {
			continue
		}
		line = strings.TrimPrefix(line, "openuem-agent:")
		line = strings.TrimPrefix(line, "openuem-updater:")
		logEntry := agents_views.LogEntry{}
		if !strings.Contains(line, "[") {
			continue
		}

		logEntry.Date = strings.TrimSpace(line[:strings.Index(line, " [")])

		switch {
		case strings.Contains(line, "[ERROR]"):
			logEntry.Category = "ERROR"
		case strings.Contains(line, "[INFO]"):
			logEntry.Category = "INFO"
		case strings.Contains(line, "[WARNING]"):
			logEntry.Category = "WARNING"
		case strings.Contains(line, "[DEBUG]"):
			logEntry.Category = "DEBUG"
		}

		if category != "" && logEntry.Category != category {
			continue
		}

		if !strings.Contains(line, "]: ") {
			continue
		}

		line = line[(strings.Index(line, "]: ") + 3):]
		logEntry.Text = line
		logEntries = append(logEntries, logEntry)
	}
	return logEntries
}

func (h *Handler) AgentSettings(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), true))
	}

	currentAgent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent", err.Error()), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	if c.Request().Method == "POST" {
		s := openuem_nats.AgentSetting{}

		s.DebugMode = false
		if c.FormValue("debug-mode") != "" {
			s.DebugMode = true
		}

		s.SFTPService = false
		if c.FormValue("sftp-service") != "" {
			s.SFTPService = true
		}

		s.RemoteAssistance = false
		if c.FormValue("remote-assistance-service") != "" {
			s.RemoteAssistance = true
		}

		s.SFTPPort = c.FormValue("sftp-port")
		if s.SFTPPort != "" {
			portNumber, err := strconv.Atoi(s.SFTPPort)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.port_must_be_a_number"), true))
			}

			if portNumber < 0 || portNumber > 65535 {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.port_is_not_valid"), true))
			}
		}

		s.VNCProxyPort = c.FormValue("vnc-proxy-port")
		if s.VNCProxyPort != "" {
			portNumber, err := strconv.Atoi(s.VNCProxyPort)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.port_must_be_a_number"), true))
			}

			if portNumber < 0 || portNumber > 65535 {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.port_is_not_valid"), true))
			}
		}

		data, err := json.Marshal(s)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.settings_data_error"), true))
		}

		if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "nats.not_connected"), false))
		}

		err = h.NATSConnection.Publish("agent.settings."+agentId, data)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.settings_nats_error", err.Error()), true))
		}

		a, err := h.Model.SaveAgentSettings(agentId, s)
		if err != nil {
			errMessage := err.Error()
			// Rollback
			s := openuem_nats.AgentSetting{}
			s.DebugMode = currentAgent.DebugMode
			s.RemoteAssistance = currentAgent.RemoteAssistance
			s.SFTPService = currentAgent.SftpService
			s.SFTPPort = currentAgent.SftpPort
			s.VNCProxyPort = currentAgent.VncProxyPort

			data, err := json.Marshal(s)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.settings_data_error"), true))
			}

			err = h.NATSConnection.Publish("agent.settings."+agentId, data)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.settings_nats_error", err.Error()), true))
			}

			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.settings_nats_error", errMessage), true))
		}

		return RenderView(c, agents_views.AgentsIndex("| Agents", agents_views.AgentSettings(c, a, i18n.T(c.Request().Context(), "agents.settings_success"), "", refreshTime, commonInfo), commonInfo))
	}

	return RenderView(c, agents_views.AgentsIndex("| Agents", agents_views.AgentSettings(c, currentAgent, "", "", refreshTime, commonInfo), commonInfo))
}
