package handlers

import (
	"log"
	"strconv"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

// ListTenantMembers shows the members (users) assigned to a tenant with their roles
func (h *Handler) ListTenantMembers(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"), false))
	}

	members, err := h.Model.GetTenantUsersWithRoles(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	currentUsername := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")

	return RenderView(c, admin_views.TenantMembersIndex(" | Members",
		admin_views.TenantMembers(c, members, "", "", agentsExists, serversExists, commonInfo, currentUsername),
		commonInfo))
}

// AddTenantMember looks up a user by email or username and assigns them to the tenant
func (h *Handler) AddTenantMember(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"), true))
	}

	identifier := c.FormValue("identifier")
	if identifier == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "members.identifier_required"), true))
	}

	role := c.FormValue("role")
	if role == "" {
		role = "user"
	}

	// Try to find user by username first, then by email
	userID := identifier
	exists, err := h.Model.UserExists(identifier)
	if err != nil || !exists {
		// Try by email
		userID = h.Model.GetUserIDByEmail(identifier)
		if userID == "" {
			return h.listTenantMembersWithError(c, commonInfo, identifier,
				i18n.T(c.Request().Context(), "members.user_not_found"))
		}
	}

	// Check if user is already a member
	hasAccess, _ := h.Model.UserHasAccessToTenant(userID, tenantID)
	if hasAccess {
		return h.listTenantMembersWithError(c, commonInfo, identifier,
			i18n.T(c.Request().Context(), "members.already_member"))
	}

	err = h.Model.AssignUserToTenant(userID, tenantID, models.UserTenantRole(role), false)
	if err != nil {
		log.Printf("[ERROR]: could not add member to tenant: %v", err)
		return h.listTenantMembersWithError(c, commonInfo, identifier, err.Error())
	}

	return h.ListTenantMembers(c)
}

// listTenantMembersWithError re-renders the members view with an error message
func (h *Handler) listTenantMembersWithError(c echo.Context, commonInfo *partials.CommonInfo, identifier, errMsg string) error {
	tenantID, _ := strconv.Atoi(commonInfo.TenantID)
	members, _ := h.Model.GetTenantUsersWithRoles(tenantID)
	agentsExists, _ := h.Model.AgentsExists(commonInfo)
	serversExists, _ := h.Model.ServersExists()
	currentUsername := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")

	return RenderView(c, admin_views.TenantMembersIndex(" | Members",
		admin_views.TenantMembers(c, members, identifier, errMsg, agentsExists, serversExists, commonInfo, currentUsername),
		commonInfo))
}

// RemoveTenantMember removes a user from the current tenant
func (h *Handler) RemoveTenantMember(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"), true))
	}

	userID := c.Param("uid")
	if userID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.user_not_found"), true))
	}

	// Prevent admin from removing themselves
	currentUsername := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if userID == currentUsername {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "members.cannot_remove_self"), true))
	}

	err = h.Model.RemoveUserFromTenant(userID, tenantID)
	if err != nil {
		log.Printf("[ERROR]: could not remove member from tenant: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.ListTenantMembers(c)
}

// UpdateTenantMemberRole updates the role of a user within the current tenant
func (h *Handler) UpdateTenantMemberRole(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"), true))
	}

	userID := c.Param("uid")
	if userID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "users.user_not_found"), true))
	}

	role := c.FormValue("role")
	if role != "admin" && role != "operator" && role != "user" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.invalid_role"), true))
	}

	// Prevent admin from demoting themselves
	currentUsername := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if userID == currentUsername && role != "admin" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "members.cannot_demote_self"), true))
	}

	err = h.Model.UpdateUserTenantRole(userID, tenantID, models.UserTenantRole(role))
	if err != nil {
		log.Printf("[ERROR]: could not update member role: %v", err)
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return h.ListTenantMembers(c)
}
