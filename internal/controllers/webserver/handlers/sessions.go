package handlers

import (
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) ListSessions(c echo.Context, successMessage string) error {
	var err error

	errMessage := ""

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

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

	agentsExists, err := h.Model.AgentsExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.SessionsIndex(" | Sessions", admin_views.Sessions(c, p, h.SessionManager, l, h.Version, latestServerRelease.Version, s, successMessage, errMessage, h.SessionManager.Manager.Codec, agentsExists, serversExists)))
}

func (h *Handler) SessionDelete(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return RenderError(c, partials.ErrorMessage("no token was found in request", true))
	}

	return RenderConfirm(c, partials.ConfirmDelete(c, i18n.T(c.Request().Context(), "confirm.session_delete"), "/admin/sessions/", "/admin/sessions/"+token))
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
