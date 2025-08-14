package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/auth"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) Logout(c echo.Context) error {
	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_get_settings", err.Error()), true))
	}

	uid, ok := h.SessionManager.Manager.Get(c.Request().Context(), "uid").(string)
	if !ok || len(uid) == 0 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_get_user_id"), true))
	}

	u, err := h.Model.GetUserById(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "authentication.could_not_get_user_info", err.Error()), true))
	}

	if err := h.SessionManager.Manager.Destroy(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if u.Openid {
		logoutURL := ""
		redirecURI := fmt.Sprintf("https://%s:%s", h.ServerName, h.ConsolePort)
		if h.ReverseProxyAuthPort != "" {
			redirecURI = fmt.Sprintf("https://%s", strings.TrimSuffix(c.Request().Referer(), "/"))
		}

		switch settings.OIDCProvider {
		case auth.AUTHELIA:
			logoutURL = fmt.Sprintf("%s/logout?rd=%s", settings.OIDCIssuerURL, redirecURI)
		case auth.AUTHENTIK:
			logoutURL = fmt.Sprintf("%send-session/", settings.OIDCIssuerURL)
		case auth.KEYCLOAK:
			logoutURL = fmt.Sprintf("%s/protocol/openid-connect/logout?client_id=%s&post_logout_redirect_uri=%s", settings.OIDCIssuerURL, settings.OIDCClientID, redirecURI)
		case auth.ZITADEL:
			logoutURL = fmt.Sprintf("%s/oidc/v1/end_session?client_id=%s&post_logout_redirect_uri=%s", settings.OIDCIssuerURL, settings.OIDCClientID, redirecURI)
		}

		c.Response().Header().Set("HX-Redirect", logoutURL)
		return c.String(http.StatusFound, "")
	}

	return h.Login(c)
}
