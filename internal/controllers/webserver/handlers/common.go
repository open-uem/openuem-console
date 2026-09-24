package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/openuem-console/internal/authz"
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
		IsAdmin:        false,
		IsProfile:      strings.Contains(c.Request().URL.String(), "profiles"),
		IsTask:         strings.Contains(c.Request().URL.String(), "tasks"),
		CSRFToken:      csrfToken,
	}

	scope, ok := c.Get(accessScopeContextKey).(*authz.AccessScope)
	if !ok || scope == nil {
		return nil, echo.NewHTTPError(http.StatusForbidden, i18n.T(c.Request().Context(), "authentication.not_authenticated"))
	}
	info.IsAdmin = scope.IsAdmin
	info.Scope = scope

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

	info.Tenants, err = h.Model.GetTenantsForScope(scope)
	if err != nil {
		return nil, err
	}

	if tenantID == "" {
		if info.IsAdmin || info.IsProfile || info.IsTask {
			info.TenantID = "-1"
			info.SiteID = "-1"
			return &info, nil
		}
		tenant, err = h.Model.GetDefaultTenantForScope(scope)
		if err != nil {
			return nil, err
		}
		info.TenantID = strconv.Itoa(tenant.ID)
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return nil, err
		}

		tenant, err = h.Model.GetTenantByIDForScope(id, scope)
		if err != nil {
			tenant, err = h.Model.GetDefaultTenantForScope(scope)
			if err != nil {
				return nil, err
			}
			info.TenantID = strconv.Itoa(tenant.ID)
		} else {
			info.TenantID = tenantID
		}
	}

	info.Sites, err = h.Model.GetAssociatedSitesForScope(tenant, scope)
	if err != nil {
		return nil, err
	}

	if siteID != "" {
		id, err := strconv.Atoi(siteID)
		if err != nil {
			return nil, err
		}

		_, err = h.Model.GetSiteByIdForScope(tenant.ID, id, scope)
		if err != nil {
			s, err := h.Model.GetDefaultSiteForScope(tenant, scope)
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
		s, err := h.Model.GetDefaultSiteForScope(tenant, scope)
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

	// is turnstile enabled
	tsSiteKey, tsSecretKey, err := h.Model.GetTurnstileSettings()
	if err != nil {
		return nil, errors.New(i18n.T(c.Request().Context(), "settings.turnstile_could_not_get_settings", err))
	}

	info.IsTurnstileEnabled = tsSecretKey != "" && tsSiteKey != ""

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
