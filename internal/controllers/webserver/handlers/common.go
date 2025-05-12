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
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) GetCommonInfo(c echo.Context) (*partials.CommonInfo, error) {
	var err error
	var tenant *ent.Tenant

	info := partials.CommonInfo{
		SM:             h.SessionManager,
		CurrentVersion: h.Version,
		Translator:     views.GetTranslatorForDates(c),
		IsAdmin:        strings.Contains(c.Request().URL.String(), "admin"),
	}

	latestRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return nil, err
	}

	info.LatestVersion = latestRelease.Version

	tenantID := c.Param("tenant")
	siteID := c.Param("site")

	info.Tenants, err = h.Model.GetTenants()
	if err != nil {
		return nil, err
	}

	if tenantID == "" {
		if info.IsAdmin {
			info.TenantID = "-1"
			info.SiteID = "-1"
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
		} else {
			info.SiteID = siteID
		}
	} else {
		if len(info.Sites) == 0 {
			info.SiteID = strconv.Itoa(info.Sites[0].ID)
		} else {
			info.SiteID = "-1"
		}
	}

	info.DetectRemoteAgents, err = h.Model.GetDefaultDetectRemoteAgents(info.TenantID)
	if err != nil {
		return nil, errors.New(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"))
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
