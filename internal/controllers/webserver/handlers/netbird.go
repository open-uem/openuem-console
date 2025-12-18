package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/computers_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/utils"
)

func (h *Handler) NetbirdSettings(c echo.Context) error {
	var err error
	var successMessage string

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID := -1
	tID := c.Param("tenant")
	if tID != "" {
		tenantID, err = strconv.Atoi(tID)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
		}
		commonInfo.TenantID = tID
	} else {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.settings_empty_tenant"), true))
	}

	if c.Request().Method == "POST" {
		managementURL := c.FormValue("netbird-management-url")
		accessToken := c.FormValue("netbird-access-token")

		if accessToken == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.token_must_be_set"), true))
		}

		// if empty URL let's use the NetBird Cloud API
		if managementURL == "" {
			managementURL = "https://api.netbird.io"
		}

		if err := h.Model.SaveNetbirdSettings(tenantID, managementURL, accessToken); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.settings_not_saved", err.Error()), true))
		}

		successMessage = i18n.T(c.Request().Context(), "netbird.settings_saved")
	}

	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.NetbirdSettingsIndex(" | NetBird Settings", admin_views.NetbirdSettings(c, settings, agentsExists, serversExists, commonInfo, h.GetAdminTenantName(commonInfo), successMessage), commonInfo))
}

func (h *Handler) Netbird(c echo.Context, successMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")

	if agentID == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	// Try to get info using NATS refresh
	msg, err := h.NATSConnection.Request("agent.netbird.refresh."+agentID, nil, 10*time.Second)
	if err == nil {
		result := nats.Netbird{}
		if err := json.Unmarshal(msg.Data, &result); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_decode_response", err.Error()), true))
		}

		if result.Error == "" {
			if err := h.Model.SaveNetbirdInfo(agentID, result); err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_save_data", result.Error), true))
			}
		}
	}

	// Get data from database
	agent, err := h.Model.GetAgentNetBirdById(agentID, commonInfo)
	if err != nil || agent.Edges.Netbird == nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_doesnt_support_netbird"), true))
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
		if ent.IsNotFound(err) {
			settings = &ent.NetbirdSettings{}
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
		}
	}

	if settings.AccessToken == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.token_empty"), true))
	}

	ng, err := getGroupsFromNetbirdAPI(settings.ManagementURL, settings.AccessToken)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_groups", err.Error()), true))
	}

	netbird := settings.AccessToken != ""

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.Netbird(c, p, agent, ng, higherVersion, confirmDelete, successMessage, commonInfo, currentTenant, currentSite, allTenants, allSites, netbird), commonInfo))
}

func (h *Handler) NetbirdInstall(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	msg, err := h.NATSConnection.Request("agent.netbird.install."+agentID, nil, 10*time.Minute)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.install_request_failed", err.Error()), true))
	}

	result := nats.Netbird{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_decode_response", err.Error()), true))
	}

	if result.Error != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.install_request_failed", result.Error), true))
	}

	if err := h.Model.SaveNetbirdInfo(agentID, result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_save_data", result.Error), true))
	}

	successMessage := i18n.T(c.Request().Context(), "netbird.install_request_succeeded")

	return h.Netbird(c, successMessage)
}

func (h *Handler) NetbirdRegister(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	allowExtraDNSLabels := false
	if c.FormValue("allow-extra-dns-labels") == "on" {
		allowExtraDNSLabels = true
	}

	p, err := c.FormParams()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_parse_groups", err.Error()), true))
	}

	groups := ""
	if len(p["groups[]"]) > 0 {
		groupIDs := []string{}
		for _, g := range p["groups[]"] {
			tmp := strings.Split(g, "-")
			if len(tmp) > 1 {
				groupIDs = append(groupIDs, fmt.Sprintf(`"%s"`, tmp[1]))
			}
		}
		groups = strings.Join(groupIDs, ",")
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}

	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			settings = &ent.NetbirdSettings{}
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
		}
	}

	if settings.AccessToken == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.token_empty"), true))
	}

	setupKeyID, setupKey, err := utils.CreateNetBirdOneOffSetupKeyAPI(settings.ManagementURL, agentID, groups, allowExtraDNSLabels, settings.AccessToken)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}

	defer func() {
		if err := utils.DeleteNetBirdOneOffSetupKeyAPI(settings.ManagementURL, setupKeyID, settings.AccessToken); err != nil {
			log.Printf("[ERROR]: could not delete one-off key using Netbird API, reason: %v", err)
		}
	}()

	request := nats.NetbirdSettings{
		ManagementURL: settings.ManagementURL,
		OneOffKey:     setupKey,
	}

	data, err := json.Marshal(request)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_create_request", err.Error()), true))
	}

	msg, err := h.NATSConnection.Request("agent.netbird.register."+agentID, data, 1*time.Minute)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.register_request_failed", err.Error()), true))
	}

	result := nats.Netbird{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_decode_response", err.Error()), true))
	}

	if result.Error != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.register_request_failed", result.Error), true))
	}

	if err := h.Model.SaveNetbirdInfo(agentID, result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_save_data", result.Error), true))
	}

	successMessage := i18n.T(c.Request().Context(), "netbird.register_request_succeeded")

	return h.Netbird(c, successMessage)
}

func (h *Handler) NetbirdUninstall(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentId := c.Param("uuid")

	agent, err := h.Model.GetAgentNetBirdById(agentId, commonInfo)
	if err != nil || agent.Edges.Netbird == nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), true))
	}

	msg, err := h.NATSConnection.Request("agent.netbird.uninstall."+agentId, nil, 10*time.Minute)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.uninstall_failed", err.Error()), true))
	}

	result := nats.Netbird{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_decode_response", err.Error()), true))
	}

	if result.Error != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.uninstall_failed", result.Error), true))
	}

	if agent.Edges.Netbird.IP != "" {
		if err := h.NetbirdDeletePeer(c, true); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.delete_peer_failed", result.Error), true))
		}
	}

	if err := h.Model.SetNetbirdAsUninstalled(agentId); err != nil {
		log.Printf("[INFO]: could not update NetBird data in the database, reason: %v", err)
	}

	successMessage := i18n.T(c.Request().Context(), "netbird.uninstall_succeeded")

	return h.Netbird(c, successMessage)
}

func (h *Handler) NetbirdSwitchProfile(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	profile := c.FormValue("profile")
	if profile == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.profile_empty"), true))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}

	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			settings = &ent.NetbirdSettings{}
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
		}
	}

	data, err := json.Marshal(nats.NetbirdSettings{
		Profile:       profile,
		ManagementURL: settings.ManagementURL,
	})

	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_create_request", err.Error()), true))
	}

	msg, err := h.NATSConnection.Request("agent.netbird.switchprofile."+agentID, data, 2*time.Minute)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.profile_switch_failed", err.Error()), true))
	}

	result := nats.Netbird{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_decode_response", err.Error()), true))
	}

	if result.Error != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.profile_switch_failed", result.Error), true))
	}

	if err := h.Model.SaveNetbirdInfo(agentID, result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_save_data", result.Error), true))
	}

	successMessage := i18n.T(c.Request().Context(), "netbird.profile_switched")

	return h.Netbird(c, successMessage)
}

func (h *Handler) NetbirdRefresh(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	msg, err := h.NATSConnection.Request("agent.netbird.refresh."+agentID, nil, 5*time.Minute)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.refresh_failed", err.Error()), true))
	}

	result := nats.Netbird{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_decode_response", err.Error()), true))
	}

	if result.Error != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.refresh_failed", result.Error), true))
	}

	if err := h.Model.SaveNetbirdInfo(agentID, result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_save_data", result.Error), true))
	}

	successMessage := i18n.T(c.Request().Context(), "netbird.info_refreshed")

	return h.Netbird(c, successMessage)
}

func (h *Handler) NetbirdConnect(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}

	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			settings = &ent.NetbirdSettings{}
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
		}
	}

	request := nats.NetbirdSettings{
		ManagementURL: settings.ManagementURL,
	}

	data, err := json.Marshal(request)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_create_request", err.Error()), true))
	}

	msg, err := h.NATSConnection.Request("agent.netbird.up."+agentID, data, 30*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.connect_failed", err.Error()), true))
	}

	result := nats.Netbird{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_decode_response", err.Error()), true))
	}

	if result.Error != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.connect_failed", result.Error), true))
	}

	if err := h.Model.SaveNetbirdInfo(agentID, result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_save_data", result.Error), true))
	}

	successMessage := i18n.T(c.Request().Context(), "netbird.connect_success")

	return h.Netbird(c, successMessage)
}

func (h *Handler) NetbirdDisconnect(c echo.Context, successMessage string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if _, err := h.Model.GetAgentById(agentID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}

	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			settings = &ent.NetbirdSettings{}
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
		}
	}

	request := nats.NetbirdSettings{
		ManagementURL: settings.ManagementURL,
	}

	data, err := json.Marshal(request)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_create_request", err.Error()), true))
	}

	msg, err := h.NATSConnection.Request("agent.netbird.down."+agentID, data, 5*time.Minute)
	if err != nil {
		if strings.Contains(err.Error(), "no responders") {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.agent_offline"), true))
		}
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.disconnect_failed", err.Error()), true))
	}

	result := nats.Netbird{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_decode_response", err.Error()), true))
	}

	if result.Error != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.disconnect_failed", result.Error), true))
	}

	if err := h.Model.SaveNetbirdInfo(agentID, result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_save_data", result.Error), true))
	}

	if successMessage == "" {
		successMessage = i18n.T(c.Request().Context(), "netbird.disconnect_success")
	}

	return h.Netbird(c, successMessage)
}

func getGroupsFromNetbirdAPI(managementURL string, token string) ([]nats.NetBirdGroups, error) {

	url := fmt.Sprintf("%s/api/groups", managementURL)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	groups := []nats.NetBirdGroups{}
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (h *Handler) NetbirdDeletePeer(c echo.Context, comingFromUninstall bool) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	agent, err := h.Model.GetAgentNetBirdById(agentID, commonInfo)

	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}

	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			settings = &ent.NetbirdSettings{}
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
		}
	}

	if settings.AccessToken == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.token_empty"), true))
	}

	ip := agent.Edges.Netbird.IP
	if ip == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.empty_ip"), true))
	}

	ipElements := strings.Split(ip, "/")
	if len(ipElements) != 2 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.wrong_ip_format"), true))
	}

	peerID, err := utils.GetMyNetBirdPeerID(ipElements[0], settings.ManagementURL, settings.AccessToken)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_peer_id", err.Error()), true))
	}

	if err := utils.DeleteNetBirdPeer(peerID, settings.ManagementURL, settings.AccessToken); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_delete_peer", err.Error()), true))
	}

	if comingFromUninstall {
		return nil
	}
	return h.NetbirdDisconnect(c, i18n.T(c.Request().Context(), "netbird.peer_deleted"))
}
