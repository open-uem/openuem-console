package handlers

import (
	"net/http"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/login_views"
)

func (h *Handler) Login(c echo.Context) error {
	// if accidentally we disable the use of certificates this allows us to reenable it again
	if h.ReenableCertAuth {
		if err := h.Model.ReEnableCertificatesAuth(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_reenable_certs", err.Error()))
		}
	}

	settings, err := h.Model.GetAuthenticationSettings()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, i18n.T(c.Request().Context(), "authentication.could_not_get_settings"))
	}

	return RenderLogin(c, login_views.LoginIndex(login_views.Login(settings)))
}
