package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent/user"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) UserPermissions(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	uid := c.Param("uid")
	u, err := h.Model.GetUserWithPermissions(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	tenants, err := h.Model.GetTenants()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	sites, err := h.Model.GetAllSitesForPermissions()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	agents, err := h.Model.GetAllAgentsForPermissions(300)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, admin_views.UsersIndex(" | Permissions", admin_views.UserPermissions(c, u, tenants, sites, agents, "", "", commonInfo), commonInfo))
}

func (h *Handler) SaveUserPermissions(c echo.Context) error {
	if err := c.Request().ParseForm(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	uid := c.Param("uid")
	role := user.ConsoleRole(c.FormValue("console_role"))
	if role != user.ConsoleRoleAdmin && role != user.ConsoleRoleCustom {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid console role")
	}

	tenantIDs := make([]int, 0)
	for _, raw := range c.Request().Form["tenant_ids"] {
		id, err := strconv.Atoi(raw)
		if err == nil {
			tenantIDs = append(tenantIDs, id)
		}
	}
	siteIDs := make([]int, 0)
	for _, raw := range c.Request().Form["site_ids"] {
		id, err := strconv.Atoi(raw)
		if err == nil {
			siteIDs = append(siteIDs, id)
		}
	}
	agentIDs := append([]string{}, c.Request().Form["agent_ids"]...)

	if err := h.Model.SaveUserPermissions(uid, role, tenantIDs, siteIDs, agentIDs); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}
	u, err := h.Model.GetUserWithPermissions(uid)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}
	tenants, _ := h.Model.GetTenants()
	sites, _ := h.Model.GetAllSitesForPermissions()
	agents, _ := h.Model.GetAllAgentsForPermissions(300)
	return RenderView(c, admin_views.UsersIndex(" | Permissions", admin_views.UserPermissions(c, u, tenants, sites, agents, "Permissions saved", "", commonInfo), commonInfo))
}
