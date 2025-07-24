package handlers

import (
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/computers_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) Nickname(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderView(c, computers_views.InventoryIndex(" | Inventory", partials.Error(c, "an error occurred getting uuid param", "Computer", partials.GetNavigationUrl(commonInfo, "/computers"), commonInfo), commonInfo))
	}

	nickname := c.FormValue("nickname")
	if nickname == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.nickname_cannot_be_empty"), true))
	}

	if err := h.Model.SaveNickname(agentID, nickname, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.nickname_not_saved", err.Error()), true))
	}

	return RenderView(c, partials.EndpointName(agentID, nickname, commonInfo))
}
