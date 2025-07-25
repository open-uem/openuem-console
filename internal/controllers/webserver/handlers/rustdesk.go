package handlers

import (
	"log"
	"strconv"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) RustDeskSettings(c echo.Context) error {
	var err error
	var successMessage string

	rustdeskSettings := &ent.RustDesk{}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}

	if c.Request().Method == "POST" {
		rendezvousServer := c.FormValue("rustdesk-rendezvous-server")
		relayServer := c.FormValue("rustdesk-relay-server")
		key := c.FormValue("rustdesk-key")
		apiServer := c.FormValue("rustdesk-api-server")

		if (rendezvousServer != "" || relayServer != "" || apiServer != "") && key == "" {
			log.Println("key error")
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.key_must_be_set"), true))
		}

		if err := h.Model.SaveRustDeskSettings(tenantID, rendezvousServer, relayServer, key, apiServer); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.settings_not_saved", err.Error()), true))
		}

		successMessage = i18n.T(c.Request().Context(), "rustdesk.settings_saved")
	}

	settings, err := h.Model.GetRustDeskSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_get_rustdesk_settings", err.Error()), true))
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

	return RenderView(c, admin_views.SMTPSettingsIndex(" | RustDesk Settings", admin_views.RustDeskSettings(c, rustdeskSettings, agentsExists, serversExists, commonInfo, h.GetAdminTenantName(commonInfo), successMessage), commonInfo))
}
