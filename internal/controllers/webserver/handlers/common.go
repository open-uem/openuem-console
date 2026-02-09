package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) GetCommonInfo(c echo.Context) (*partials.CommonInfo, error) {
	var err error
	var tenant *ent.Tenant

	csrfToken, ok := c.Get("csrf").(string)
	if c.Request().Method != "GET" && (!ok || csrfToken == "") {
		return nil, errors.New("could not find CSRF token")
	}

	info := partials.CommonInfo{
		SM:             h.SessionManager,
		CurrentVersion: h.Version,
		Translator:     views.GetTranslatorForDates(c),
		IsAdmin:        strings.Contains(c.Request().URL.String(), "admin"),
		IsProfile:      strings.Contains(c.Request().URL.String(), "profiles"),
		CSRFToken:      csrfToken,
	}

	if strings.Contains(c.Request().URL.String(), "computers") && !strings.HasSuffix(c.Request().URL.String(), "computers") {
		info.IsComputer = true
	}

	// check if we're running in Docker or no server updater info is stored
	allUpdateServers, err := h.Model.GetAllUpdateServers(filters.UpdateServersFilter{})
	if err != nil {
		return nil, err
	}
	info.IsDocker = len(allUpdateServers) == 0

	latestRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return nil, err
	}

	info.LatestVersion = latestRelease.Version

	tenantID := c.Param("tenant")
	siteID := c.Param("site")

	// Get username for tenant filtering
	username := h.SessionManager.Manager.GetString(c.Request().Context(), "uid")
	if username != "" {
		// Only show tenants the user has access to
		userTenants, err := h.Model.GetTenantsForUser(username)
		if err == nil {
			info.Tenants = userTenants
		}
	} else {
		info.Tenants = []*ent.Tenant{}
	}

	if tenantID == "" {
		if info.IsAdmin {
			info.TenantID = "-1"
			info.SiteID = "-1"
			// Load branding settings for admin pages
			info.Branding, _ = h.Model.GetOrCreateBranding()
			return &info, nil
		}
		tenant, err = h.Model.GetDefaultTenant()
		if err != nil {
			return nil, err
		}
		info.TenantID = strconv.Itoa(tenant.ID)
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return nil, err
		}

		tenant, err = h.Model.GetTenantByID(id)
		if err != nil {
			tenant, err = h.Model.GetDefaultTenant()
			if err != nil {
				return nil, err
			}
			info.TenantID = strconv.Itoa(tenant.ID)
		} else {
			info.TenantID = tenantID
		}
	}

	info.Sites, err = h.Model.GetAssociatedSites(tenant)
	if err != nil {
		return nil, err
	}

	if siteID != "" {
		id, err := strconv.Atoi(siteID)
		if err != nil {
			return nil, err
		}

		_, err = h.Model.GetSiteById(tenant.ID, id)
		if err != nil {
			s, err := h.Model.GetDefaultSite(tenant)
			if err != nil {
				return nil, err
			}
			info.SiteID = strconv.Itoa(s.ID)
			info.ProfileSiteID = info.SiteID
		} else {
			info.SiteID = siteID
			info.ProfileSiteID = info.SiteID
		}
	} else {
		s, err := h.Model.GetDefaultSite(tenant)
		if err != nil {
			return nil, err
		}
		info.ProfileSiteID = strconv.Itoa(s.ID)

		if len(info.Sites) != 0 {
			info.SiteID = "-1"
		}
	}

	info.DetectRemoteAgents, err = h.Model.GetDefaultDetectRemoteAgents(info.TenantID)
	if err != nil {
		return nil, errors.New(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"))
	}

	// Load branding settings
	info.Branding, _ = h.Model.GetOrCreateBranding()

	// Multi-tenancy: Populate additional user/tenant context
	// username already defined earlier for tenant filtering
	if username != "" {
		// Check if user is admin in main tenant
		info.IsMainTenantAdmin, _ = h.Model.IsMainTenantAdmin(username)

		// Get user's role in current tenant
		info.UserRole, _ = h.GetCurrentUserTenantRole(c)

		// Get accessible tenants
		info.AccessibleTenants, _ = h.GetUserAccessibleTenants(c)

		// Check if current tenant is main tenant
		if tenant != nil {
			info.CurrentTenantIsMain, _ = h.Model.IsMainTenant(tenant.ID)
		}
	}

	return &info, nil
}

func (h *Handler) GetAdminTenantName(commonInfo *partials.CommonInfo) string {
	tenantName := ""
	if commonInfo.TenantID != "-1" {
		tenantID, err := strconv.Atoi(commonInfo.TenantID)
		if err != nil {
			return ""
		}

		t, err := h.Model.GetTenantByID(tenantID)
		if err != nil {
			return ""
		}
		tenantName = t.Description
	}
	return tenantName
}

func (h *Handler) GetAdminSiteName(commonInfo *partials.CommonInfo) string {
	siteName := ""
	if commonInfo.TenantID != "-1" {
		tenantID, err := strconv.Atoi(commonInfo.TenantID)
		if err != nil {
			return ""
		}

		siteID, err := strconv.Atoi(commonInfo.SiteID)
		if err != nil {
			return ""
		}

		s, err := h.Model.GetSiteById(tenantID, siteID)
		if err != nil {
			return ""
		}

		siteName = s.Description
	}
	return siteName
}
