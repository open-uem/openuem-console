package handlers

import (
	"errors"
	"strconv"

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
	}

	tenantID := c.Param("tenant")
	siteID := c.Param("site")

	info.Tenants, err = h.Model.GetTenants()
	if err != nil {
		return nil, err
	}

	if tenantID == "" {
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

	if siteID == "" {
		s, err := h.Model.GetDefaultSite(tenant)
		if err != nil {
			return nil, err
		}
		info.SiteID = strconv.Itoa(s.ID)
	} else {
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
	}

	latestRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return nil, err
	}

	info.LatestVersion = latestRelease.Version

	info.DetectRemoteAgents, err = h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return nil, errors.New(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"))
	}

	return &info, nil
}
