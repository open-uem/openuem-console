package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/views/admin_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) ListSites(c echo.Context, successMessage, errMessage string, confirmDelete bool) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.tenant_cannot_be_empty"), true))
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.site_cannot_be_empty"), true))
	}

	f := filters.SiteFilter{}

	nameFilter := c.FormValue("filterByName")
	if nameFilter != "" {
		f.Name = nameFilter
	}

	domainFilter := c.FormValue("filterByDomain")
	if domainFilter != "" {
		f.Domain = domainFilter
	}

	createdFrom := c.FormValue("filterByCreatedDateFrom")
	if createdFrom != "" {
		f.CreatedFrom = createdFrom
	}
	createdTo := c.FormValue("filterByCreatedDateTo")
	if createdTo != "" {
		f.CreatedTo = createdTo
	}

	modifiedFrom := c.FormValue("filterByModifiedDateFrom")
	if modifiedFrom != "" {
		f.ModifiedFrom = modifiedFrom
	}
	modifiedTo := c.FormValue("filterByModifiedDateTo")
	if modifiedTo != "" {
		f.ModifiedTo = modifiedTo
	}

	filteredDefaultStatus := []string{}
	for index := range []string{"Yes", "No"} {
		value := c.FormValue(fmt.Sprintf("filterByDefaultStatus%d", index))
		if value != "" {
			filteredDefaultStatus = append(filteredDefaultStatus, value)
		}
	}
	f.DefaultOptions = filteredDefaultStatus

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	p.NItems, err = h.Model.CountAllSites(f, commonInfo.TenantID)
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	sites, err := h.Model.GetSitesByPage(p, f, commonInfo.TenantID)
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.SitesIndex(" | Sites", admin_views.Sites(c, p, f, sites, successMessage, errMessage, refreshTime, agentsExists, serversExists, confirmDelete, commonInfo, h.GetAdminTenantName(commonInfo)), commonInfo))
}

func (h *Handler) NewSite(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.tenant_cannot_be_empty"), true))
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.site_cannot_be_empty"), true))
	}

	defaultCountry, err := h.Model.GetDefaultCountry()
	if err != nil {
		return err
	}

	agentsExists, err := h.Model.AgentsExists(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	serversExists, err := h.Model.ServersExists()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return RenderView(c, admin_views.SitesIndex(" | Sites", admin_views.NewSite(c, defaultCountry, agentsExists, serversExists, commonInfo, h.GetAdminTenantName(commonInfo)), commonInfo))
}

func (h *Handler) AddSite(c echo.Context) error {
	successMessage := ""
	errMessage := ""

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_get_common_info", err.Error()), true))
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.tenant_cannot_be_empty"), true))
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.site_cannot_be_empty"), true))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_convert_to_int", commonInfo.TenantID), true))
	}

	name := c.FormValue("name")
	if name == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.name_cannot_be_empty"), true))
	}

	exists, err := h.Model.SiteNameTaken(tenantID, name)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_check_name", err.Error()), true))
	}

	if exists {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.site_name_taken", name), true))
	}

	isDefault, err := strconv.ParseBool(c.FormValue("is-default"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_convert_to_bool", err.Error()), true))
	}

	domain := c.FormValue("domain")
	if domain == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.domain_cannot_be_empty"), true))
	}

	err = h.Model.AddSite(tenantID, name, isDefault, domain)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.new_error"), true))
	}

	successMessage = i18n.T(c.Request().Context(), "sites.new_success")
	return h.ListSites(c, successMessage, errMessage, false)
}

func (h *Handler) EditSite(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_get_common_info", err.Error()), true))
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.tenant_cannot_be_empty"), true))
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.site_cannot_be_empty"), true))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_convert_to_int", commonInfo.TenantID), true))
	}

	siteID, err := strconv.Atoi(commonInfo.SiteID)
	if err != nil {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.could_not_convert_site_to_int", commonInfo.SiteID), false)
	}

	s, err := h.Model.GetSiteById(tenantID, siteID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.site_not_found", err.Error()), true))
	}

	if c.Request().Method == "POST" {
		name := c.FormValue("name")
		if name == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.name_cannot_be_empty"), true))
		}

		domain := c.FormValue("domain")
		if domain == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.domain_cannot_be_empty"), true))
		}

		if s.Description != name {
			exists, err := h.Model.SiteNameTaken(tenantID, name)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_check_name", err.Error()), true))
			}

			if exists {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.site_name_taken", name), true))
			}
		}

		isDefault, err := strconv.ParseBool(c.FormValue("is-default"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_convert_to_bool", err.Error()), true))
		}

		if err := h.Model.UpdateSite(tenantID, s.ID, name, domain, isDefault); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		return h.ListSites(c, i18n.T(c.Request().Context(), "sites.edit_success"), "", false)
	}

	defaultCountry, err := h.Model.GetDefaultCountry()
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

	return RenderView(c, admin_views.SitesIndex(" | Sites", admin_views.EditSite(c, s, defaultCountry, agentsExists, serversExists, commonInfo, h.GetAdminTenantName(commonInfo)), commonInfo))
}

func (h *Handler) DeleteSite(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.could_not_get_common_info", err.Error()), false)
	}

	if commonInfo.TenantID == "-1" {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.tenant_cannot_be_empty"), false)
	}

	if commonInfo.TenantID == "-1" {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.site_cannot_be_empty"), false)
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.could_not_convert_to_int", commonInfo.TenantID), false)
	}

	siteID, err := strconv.Atoi(commonInfo.SiteID)
	if err != nil {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.could_not_convert_site_to_int", commonInfo.SiteID), false)
	}

	nSites, err := h.Model.CountSites(tenantID)
	if err != nil {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.could_not_count_sites"), false)
	}

	if nSites == 1 {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.at_least_one_site"), false)
	}

	s, err := h.Model.GetSiteById(tenantID, siteID)
	if err != nil {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.could_not_find_site"), false)
	}

	if s.IsDefault {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.default_cannot_be_deleted"), false)
	}

	// Send a request to uninstall agents associated with this organization
	agents, err := h.Model.GetAgentsBySite(tenantID, siteID)
	if err != nil {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.could_not_get_agents"), false)
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "nats.not_connected"), false)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, a := range agents {
		if _, err := h.JetStream.Publish(ctx, "agent.uninstall."+a.ID, nil); err != nil {
			return h.ListSites(c, "", i18n.T(c.Request().Context(), "agents.could_not_send_request_to_uninstall"), false)
		}
	}

	// Remove the site with cascade
	if err := h.Model.DeleteSite(tenantID, siteID); err != nil {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.delete_error", err.Error()), false)
	}

	successMessage := i18n.T(c.Request().Context(), "sites.deleted")
	return h.ListSites(c, successMessage, "", false)
}

func (h *Handler) ImportSites(c echo.Context) error {
	var errors = []string{}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_get_common_info", err.Error()), false))
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.tenant_cannot_be_empty"), true))
	}

	if commonInfo.TenantID == "-1" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.site_cannot_be_empty"), true))
	}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_convert_to_int", commonInfo.TenantID), false))
	}

	// Source
	file, err := c.FormFile("csvFile")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	r := csv.NewReader(src)

	index := 1

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.import_read_error", err.Error()), false))

		}

		if len(record) != 2 {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.import_error_wrong_format", index), false))
		}

		if record[0] == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.import_required", "orgname", index), false))
		}

		if record[1] == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.import_required", "domain", index), false))
		}

		index++

		exists, err := h.Model.SiteNameTaken(tenantID, record[0])
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		if exists {
			errors = append(errors, i18n.T(c.Request().Context(), "sites.site_name_taken", record[0]))
			continue
		}

		err = h.Model.AddSite(tenantID, record[0], false, record[1])
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
	}

	if len(errors) > 0 {
		return h.ListSites(c, "", i18n.T(c.Request().Context(), "sites.import_wrong_sites", strings.Join(errors, ",")), false)
	}

	return h.ListSites(c, i18n.T(c.Request().Context(), "sites.import_success"), "", false)
}
