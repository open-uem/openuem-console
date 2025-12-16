package handlers

import (
	"encoding/json"
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
	"github.com/sethvargo/go-password/password"
)

func (h *Handler) RustDeskStart(c echo.Context) error {
	rustdeskSettings := &ent.Rustdesk{}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}

	agentId := c.Param("uuid")
	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	settings, err := h.Model.GetRustDeskSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_get_rustdesk_settings", err.Error()), true))
	}

	if len(settings) > 0 {
		rustdeskSettings = settings[0]
	}

	rd := nats.RustDesk{
		CustomRendezVousServer: rustdeskSettings.CustomRendezvousServer,
		RelayServer:            rustdeskSettings.RelayServer,
		Key:                    rustdeskSettings.Key,
		APIServer:              rustdeskSettings.APIServer,
		DirectIPAccess:         rustdeskSettings.DirectIPAccess,
		Whitelist:              rustdeskSettings.Whitelist,
	}

	randomPassword := ""
	if rustdeskSettings.UsePermanentPassword {
		randomPassword, err = password.Generate(32, 10, 0, false, true)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_generate_random_password", err.Error()), true))
		}
		rd.PermanentPassword = randomPassword
	}

	data, err := json.Marshal(rd)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_prepare_request", err.Error()), true))
	}

	msg, err := h.NATSConnection.Request("agent.rustdesk.start."+agentId, data, time.Duration(h.NATSTimeout)*time.Second)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_send_request", err.Error()), true))
	}

	result := nats.RustDeskResult{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_decode_response", err.Error()), true))
	}

	if result.Error != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.remote_error", result.Error), true))
	}

	IPAddresses := []string{}
	for _, n := range agent.Edges.Networkadapters {
		addresses := strings.Split(n.Addresses, ",")
		for _, a := range addresses {
			IPAddresses = append(IPAddresses, a)
		}
	}

	return RenderView(c, computers_views.RustDeskControl(agentId, result.RustDeskID, rustdeskSettings, IPAddresses, randomPassword, agent.IsWayland && agent.IsFlatpakRustdesk, commonInfo))
}

func (h *Handler) RustDeskStop(c echo.Context) error {
	agentId := c.Param("uuid")

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, err.Error(), "Computers", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	hasRustDeskSettings := h.Model.HasRustDeskSettings(tenantID)

	confirmDelete := c.QueryParam("delete") != ""
	p := partials.PaginationAndSort{}

	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}
	netbird := settings.AccessToken != ""

	agent, err := h.Model.GetAgentById(agentId, commonInfo)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.RemoteAssistance(c, p, agent, confirmDelete, hasRustDeskSettings, false, commonInfo, err.Error(), netbird), commonInfo))
	}

	msg, err := h.NATSConnection.Request("agent.rustdesk.stop."+agentId, nil, time.Duration(h.NATSTimeout)*time.Second)
	if err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.RemoteAssistance(c, p, agent, confirmDelete, hasRustDeskSettings, false, commonInfo, err.Error(), netbird), commonInfo))
	}

	result := nats.RustDeskResult{}
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.RemoteAssistance(c, p, agent, confirmDelete, hasRustDeskSettings, false, commonInfo, result.Error, netbird), commonInfo))
	}

	if result.Error != "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.RemoteAssistance(c, p, agent, confirmDelete, hasRustDeskSettings, false, commonInfo, result.Error, netbird), commonInfo))
	}

	return RenderView(c, computers_views.InventoryIndex(" | Inventory", computers_views.RemoteAssistance(c, p, agent, confirmDelete, hasRustDeskSettings, false, commonInfo, "", netbird), commonInfo))
}

func (h *Handler) RustDeskSettings(c echo.Context) error {
	var err error
	var successMessage string

	rustdeskSettings := &ent.Rustdesk{}

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
		commonInfo.TenantID = "-1"
	}

	if c.Request().Method == "POST" {
		rendezvousServer := c.FormValue("rustdesk-rendezvous-server")
		relayServer := c.FormValue("rustdesk-relay-server")
		key := c.FormValue("rustdesk-key")
		apiServer := c.FormValue("rustdesk-api-server")

		useDirectAccess, err := strconv.ParseBool(c.FormValue("rustdesk-direct-ip-access"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_parse_direct_ip"), true))
		}

		whitelist := c.FormValue("rustdesk-whitelist")

		usePassword, err := strconv.ParseBool(c.FormValue("rustdesk-password"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_parse_permanent"), true))
		}

		if (rendezvousServer != "" || relayServer != "" || apiServer != "") && key == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.key_must_be_set"), true))
		}

		if err := h.Model.SaveRustDeskSettings(tenantID, rendezvousServer, relayServer, key, apiServer, whitelist, useDirectAccess, usePassword); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.settings_not_saved", err.Error()), true))
		}

		successMessage = i18n.T(c.Request().Context(), "rustdesk.settings_saved")
	}

	settings := []*ent.Rustdesk{}

	if tenantID == -1 {
		settings, err = h.Model.GetGlobalRustDeskSettings()
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_get_global_rustdesk_settings", err.Error()), true))
		}
	} else {
		settings, err = h.Model.GetTenantRustDeskSettings(tenantID)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_get_rustdesk_settings", err.Error()), true))
		}
	}

	if len(settings) > 0 {
		rustdeskSettings = settings[0]
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.RustDeskSettingsIndex(" | RustDesk Settings", admin_views.RustDeskSettings(c, rustdeskSettings, agentsExists, serversExists, commonInfo, h.GetAdminTenantName(commonInfo), successMessage), commonInfo))
}

func (h *Handler) ApplyGlobalRustDeskSettings(c echo.Context) error {
	rustdeskSettings := &ent.Rustdesk{}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	settings, err := h.Model.GetGlobalRustDeskSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_get_global_rustdesk_settings", err.Error()), true))
	}

	if len(settings) > 0 {
		rustdeskSettings = settings[0]
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	successMessage := ""

	return RenderView(c, admin_views.SMTPSettingsIndex(" | RustDesk Settings", admin_views.RustDeskSettings(c, rustdeskSettings, agentsExists, serversExists, commonInfo, h.GetAdminTenantName(commonInfo), successMessage), commonInfo))
}
