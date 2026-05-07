package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/openuem-console/internal/authz"
)

const accessScopeContextKey = "access_scope"

func getScope(c echo.Context) *authz.AccessScope {
	scope, ok := c.Get(accessScopeContextKey).(*authz.AccessScope)
	if !ok {
		return nil
	}
	return scope
}

func (h *Handler) enforceRouteAuthorization(c echo.Context, scope *authz.AccessScope) error {
	path := c.Request().URL.Path

	if strings.HasPrefix(path, "/admin") && !scope.IsAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "admin permissions are required")
	}

	tenantParam := c.Param("tenant")
	if tenantParam != "" {
		tenantID, err := strconv.Atoi(tenantParam)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant")
		}
		if !scope.AllowsTenant(tenantID) {
			return echo.NewHTTPError(http.StatusForbidden, "tenant access denied")
		}
	}

	siteParam := c.Param("site")
	if siteParam != "" {
		siteID, err := strconv.Atoi(siteParam)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid site")
		}
		if !scope.AllowsSite(siteID) {
			siteTenant, stErr := h.Model.Client.Site.Query().Where(site.ID(siteID)).QueryTenant().Only(c.Request().Context())
			if stErr != nil || !scope.AllowsTenant(siteTenant.ID) {
				return echo.NewHTTPError(http.StatusForbidden, "site access denied")
			}
		}
	}

	agentID := c.Param("uuid")
	if agentID != "" && !scope.IsAdmin {
		allowed := scope.AllowsAgent(agentID)
		if !allowed {
			a, err := h.Model.Client.Agent.Query().Where(agent.ID(agentID)).WithSite(func(q *ent.SiteQuery) { q.WithTenant() }).Only(c.Request().Context())
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "agent not found")
			}
			for _, s := range a.Edges.Site {
				if scope.AllowsSite(s.ID) {
					allowed = true
					break
				}
				if s.Edges.Tenant != nil && scope.AllowsTenant(s.Edges.Tenant.ID) {
					allowed = true
					break
				}
				if allowed {
					break
				}
			}
		}
		if !allowed {
			return echo.NewHTTPError(http.StatusForbidden, "agent access denied")
		}
	}

	return nil
}
