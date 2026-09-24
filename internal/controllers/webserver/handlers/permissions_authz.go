package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/openuem-console/internal/authz"
)

var tenantAdminPathRe = regexp.MustCompile(`^/tenant/\d+/admin`)

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

	if (strings.HasPrefix(path, "/admin") || tenantAdminPathRe.MatchString(path)) && !scope.IsAdmin {
		tenantParam := c.Param("tenant")
		siteParam := c.Param("site")
		var dashboardURL string
		switch {
		case tenantParam != "" && siteParam != "":
			dashboardURL = fmt.Sprintf("/tenant/%s/site/%s/dashboard", tenantParam, siteParam)
		case tenantParam != "":
			dashboardURL = fmt.Sprintf("/tenant/%s/dashboard", tenantParam)
		default:
			dashboardURL = "/dashboard"
		}
		return c.Redirect(http.StatusFound, dashboardURL)
	}

	tenantParam := c.Param("tenant")
	if tenantParam != "" {
		tenantID, err := strconv.Atoi(tenantParam)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant")
		}
		if !scope.CanAccessTenant(tenantID) {
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
	var agentWithSite *ent.Agent
	if agentID != "" && !scope.IsAdmin {
		allowed := scope.AllowsAgent(agentID)
		if !allowed {
			a, err := h.Model.Client.Agent.Query().Where(agent.ID(agentID)).WithSite(func(q *ent.SiteQuery) { q.WithTenant() }).Only(c.Request().Context())
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "agent not found")
			}
			agentWithSite = a
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

	// Redirect /computers/:uuid[/subpath] to /tenant/:id/site/:id/computers/:uuid[/subpath]
	if agentID != "" && tenantParam == "" && strings.Contains(path, "/computers/") {
		if agentWithSite == nil {
			a, err := h.Model.Client.Agent.Query().Where(agent.ID(agentID)).WithSite(func(q *ent.SiteQuery) { q.WithTenant() }).Only(c.Request().Context())
			if err == nil {
				agentWithSite = a
			}
		}
		if agentWithSite != nil && len(agentWithSite.Edges.Site) > 0 && agentWithSite.Edges.Site[0].Edges.Tenant != nil {
			siteID := agentWithSite.Edges.Site[0].ID
			tenantID := agentWithSite.Edges.Site[0].Edges.Tenant.ID
			newPath := fmt.Sprintf("/tenant/%d/site/%d%s", tenantID, siteID, c.Request().URL.RequestURI())
			if c.Request().Header.Get("HX-Request") != "" {
				c.Response().Header().Set("HX-Redirect", newPath)
				return c.NoContent(http.StatusOK)
			}
			return c.Redirect(http.StatusFound, newPath)
		}
	}

	return nil
}
