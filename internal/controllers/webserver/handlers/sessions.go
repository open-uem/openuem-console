package handlers

import (
	"github.com/doncicuto/openuem-console/internal/views"
	"github.com/doncicuto/openuem-console/internal/views/admin_views"
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

	l := views.GetTranslatorForDates(c)

	return RenderView(c, admin_views.SessionsIndex(" | Sessions", admin_views.Sessions(c, p, h.SessionManager, l, s, successMessage, errMessage, h.SessionManager.Manager.Codec)))
}

func (h *Handler) SessionDelete(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return RenderError(c, partials.ErrorMessage("no token was found in request", true))
	}

	return RenderConfirm(c, partials.ConfirmDelete(i18n.T(c.Request().Context(), "confirm.session_delete"), "", "/admin/sessions/"+token, c.Request().Referer()))
}

func (h *Handler) SessionConfirmDelete(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return RenderError(c, partials.ErrorMessage("no token was found in request", true))
	}

	if err := h.Model.DeleteSession(token); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.ListSessions(c, i18n.T(c.Request().Context(), "success.session_delete"))
}
