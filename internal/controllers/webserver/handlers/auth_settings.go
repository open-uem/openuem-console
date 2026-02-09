package handlers

import (
	"slices"
	"strconv"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/auth"
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
		oidcProvider := c.FormValue("authentication-oidc-provider")
		oidcServer := c.FormValue("authentication-oidc-server")
		oidcClientID := c.FormValue("authentication-oidc-client-id")
		oidcRoleAdmin := c.FormValue("authentication-oidc-role-admin")
		oidcRoleOperator := c.FormValue("authentication-oidc-role-operator")
		oidcRoleUser := c.FormValue("authentication-oidc-role-user")

		useCertificates, err := strconv.ParseBool(c.FormValue("authentication-use-certificates"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_parse_use_certificates"), true))
		}

		allowRegister, err := strconv.ParseBool(c.FormValue("authentication-allow-register"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_parse_allow_register"), true))
		}

		useOIDC, err := strconv.ParseBool(c.FormValue("authentication-use-oidc"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_parse_use_oidc"), true))
		}

		usePasswd, err := strconv.ParseBool(c.FormValue("authentication-use-passwords"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_parse_use_passwords"), true))
		}

		if !useCertificates && !useOIDC {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.at_least_one_auth_method"), true))
		}

		if !useCertificates && allowRegister {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.use_certs_must_be_enabled"), true))
		}

		// if we explicitely don't want to use certs, override reenablecertauth option
		if !useCertificates {
			h.ReenableCertAuth = false
		}

		// if we explicitely don't want to use passwords, override reenablepasswdauth option
		if !usePasswd {
			h.ReenablePasswdAuth = false
		}

		if !useOIDC {
			oidcProvider = ""
			oidcServer = ""
			oidcClientID = ""
			oidcRoleAdmin = ""
			oidcRoleOperator = ""
			oidcRoleUser = ""
		}

		autoCreate, err := strconv.ParseBool(c.FormValue("authentication-oidc-auto-create"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_parse_oidc_auto_create"), true))
		}

		autoApprove, err := strconv.ParseBool(c.FormValue("authentication-oidc-auto-approve"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_parse_oidc_auto_approve"), true))
		}

		allowedProviders := []string{auth.AUTHELIA, auth.AUTHENTIK, auth.KEYCLOAK, auth.ZITADEL}
		if useOIDC && (oidcProvider == "" || (oidcProvider != "" && !slices.Contains(allowedProviders, oidcProvider))) {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.provider_not_valid"), true))
		}

		if useOIDC && oidcServer == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.oidc_url_is_required"), true))
		}

		if useOIDC && oidcClientID == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.client_id_is_required"), true))
		}

		if useOIDC && (autoCreate || autoApprove) && oidcRoleAdmin == "" && oidcRoleOperator == "" && oidcRoleUser == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.role_required"), true))
		}

		if err := h.Model.SaveAuthenticationSettings(useCertificates, allowRegister, useOIDC, oidcProvider, oidcServer, oidcClientID, oidcRoleAdmin, oidcRoleOperator, oidcRoleUser, autoCreate, autoApprove, usePasswd); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.settings_not_saved", err.Error()), true))
		}

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

	return RenderView(c, admin_views.AuthenticationSettingsIndex(" | Authentication Settings", admin_views.AuthenticationSettings(c, settings, agentsExists, serversExists, commonInfo, successMessage), commonInfo))
}
