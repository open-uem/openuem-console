package handlers

import (
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) AuthenticationSettings(c echo.Context) error {
	var err error
	var successMessage string

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if c.Request().Method == "POST" {
		// rendezvousServer := c.FormValue("rustdesk-rendezvous-server")
		// relayServer := c.FormValue("rustdesk-relay-server")
		// key := c.FormValue("rustdesk-key")
		// apiServer := c.FormValue("rustdesk-api-server")

		// useDirectAccess, err := strconv.ParseBool(c.FormValue("rustdesk-direct-ip-access"))
		// if err != nil {
		// 	return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_parse_direct_ip"), true))
		// }

		// whitelist := c.FormValue("rustdesk-whitelist")

		// usePassword, err := strconv.ParseBool(c.FormValue("rustdesk-password"))
		// if err != nil {
		// 	return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.could_not_parse_permanent"), true))
		// }

		// if (rendezvousServer != "" || relayServer != "" || apiServer != "") && key == "" {
		// 	log.Println("key error")
		// 	return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.key_must_be_set"), true))
		// }

		// if err := h.Model.SaveRustDeskSettings(tenantID, rendezvousServer, relayServer, key, apiServer, whitelist, useDirectAccess, usePassword); err != nil {
		// 	return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "rustdesk.settings_not_saved", err.Error()), true))
		// }

		successMessage = i18n.T(c.Request().Context(), "authentication.settings_saved")
	}

	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_get_settings", err.Error()), true))
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.AuthenticationSettingsIndex(" | RustDesk Settings", admin_views.AuthenticationSettings(c, settings, agentsExists, serversExists, commonInfo, successMessage), commonInfo))
}
