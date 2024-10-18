package handlers

import (
	"github.com/doncicuto/openuem-console/internal/views/config_views"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListSessions(c echo.Context, successMessage string) error {
	var err error

	errMessage := ""

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c)

	p.NItems, err = h.Model.CountAllSessions()
	if err != nil {
		errMessage = err.Error()
	}

	s, err := h.Model.GetSessionsByPage(p)
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	return renderView(c, config_views.SessionsIndex(" | Sessions", config_views.Sessions(c, s, p, successMessage, errMessage, h.SessionManager.Manager.Codec)))
}

func (h *Handler) SessionDelete(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return renderError(c, partials.ErrorMessage("no token was found in request", true))
	}

	return renderConfirm(c, partials.ConfirmDelete(i18n.T(c.Request().Context(), "confirm.session_delete"), "", "/admin/sessions/"+token))
}

func (h *Handler) SessionConfirmDelete(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return renderError(c, partials.ErrorMessage("no token was found in request", true))
	}

	if err := h.Model.DeleteSession(token); err != nil {
		return renderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.ListSessions(c, i18n.T(c.Request().Context(), "success.session_delete"))
}
