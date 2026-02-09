package handlers

import (
	"net/http"
	"strconv"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

// TenantAccessMiddleware checks if the authenticated user has access to the requested tenant
func (h *Handler) TenantAccessMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get user ID from session
		username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
		if username == "" {
			return h.Login(c)
		}

		// Get tenant ID from URL parameter
		tenantIDStr := c.Param("tenant")
		if tenantIDStr == "" || tenantIDStr == "-1" {
			// No specific tenant requested, continue
			return next(c)
		}

		tenantID, err := strconv.Atoi(tenantIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"))
		}

		// Check if user has access to this tenant
		hasAccess, err := h.Model.UserHasAccessToTenant(username, tenantID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if !hasAccess {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "tenants.no_access"))
		}

		// Store tenant access info in context for later use
		c.Set("tenant_id", tenantID)
		c.Set("user_id", username)

		return next(c)
	}
}

// TenantAdminMiddleware checks if the user is an admin in the current tenant
func (h *Handler) TenantAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get user ID from session
		username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
		if username == "" {
			return h.Login(c)
		}

		// Get tenant ID from URL parameter
		tenantIDStr := c.Param("tenant")
		if tenantIDStr == "" {
			return echo.NewHTTPError(http.StatusBadRequest, i18n.T(c.Request().Context(), "tenants.tenant_required"))
		}

		tenantID, err := strconv.Atoi(tenantIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"))
		}

		// Check if user is admin in this tenant
		isAdmin, err := h.Model.IsUserTenantAdmin(username, tenantID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if !isAdmin {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "tenants.admin_required"))
		}

		return next(c)
	}
}

// MainTenantAdminMiddleware checks if the user is an admin in the main tenant (for global settings)
// The main tenant is simply the tenant with the lowest ID
func (h *Handler) MainTenantAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get user ID from session
		username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
		if username == "" {
			return h.Login(c)
		}

		// Get main tenant (lowest ID)
		mainTenant, err := h.Model.GetMainTenant()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Check if user is admin in the main tenant
		isMainAdmin, err := h.Model.IsUserTenantAdmin(username, mainTenant.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if !isMainAdmin {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "tenants.main_admin_required"))
		}

		return next(c)
	}
}

// TenantOperatorMiddleware checks if the user is an admin OR operator in the tenant (for settings access)
func (h *Handler) TenantOperatorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get user ID from session
		username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
		if username == "" {
			return h.Login(c)
		}

		// Get tenant ID from URL parameter
		tenantIDStr := c.Param("tenant")
		if tenantIDStr == "" {
			return echo.NewHTTPError(http.StatusBadRequest, i18n.T(c.Request().Context(), "tenants.tenant_required"))
		}

		tenantID, err := strconv.Atoi(tenantIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, i18n.T(c.Request().Context(), "tenants.invalid_tenant_id"))
		}

		// Check if user is admin or operator in this tenant
		role, err := h.Model.GetUserRoleInTenant(username, tenantID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if role != models.UserTenantRoleAdmin && role != models.UserTenantRoleOperator {
			return echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "tenants.operator_required"))
		}

		return next(c)
	}
}

// GetCurrentUserTenantRole returns the role of the current user in the current tenant
func (h *Handler) GetCurrentUserTenantRole(c echo.Context) (string, error) {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return "", nil
	}

	// Get tenant ID
	tenantIDStr := c.Param("tenant")
	if tenantIDStr == "" || tenantIDStr == "-1" {
		return "", nil
	}

	tenantID, err := strconv.Atoi(tenantIDStr)
	if err != nil {
		return "", err
	}

	role, err := h.Model.GetUserRoleInTenant(username, tenantID)
	if err != nil {
		return "", err
	}

	return string(role), nil
}

// GetUserAccessibleTenants returns all tenants the current user can access
func (h *Handler) GetUserAccessibleTenants(c echo.Context) ([]*partials.TenantInfo, error) {
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username == "" {
		return nil, nil
	}

	tenants, err := h.Model.GetTenantsForUser(username)
	if err != nil {
		return nil, err
	}

	// Get the main tenant to check which one it is
	mainTenant, _ := h.Model.GetMainTenant()
	mainTenantID := 0
	if mainTenant != nil {
		mainTenantID = mainTenant.ID
	}

	result := make([]*partials.TenantInfo, 0, len(tenants))
	for _, t := range tenants {
		role, _ := h.Model.GetUserRoleInTenant(username, t.ID)
		result = append(result, &partials.TenantInfo{
			ID:          t.ID,
			Description: t.Description,
			IsDefault:   t.IsDefault,
			IsMain:      t.ID == mainTenantID,
			UserRole:    string(role),
		})
	}

	return result, nil
}
