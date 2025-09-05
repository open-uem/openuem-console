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

type NewTenant struct {
	UID     string `form:"uid" validate:"required"`
	Name    string `form:"name" validate:"required"`
	Email   string `form:"email" validate:"required,email"`
	Phone   string `form:"phone"`
	Country string `form:"country"`
}

func (h *Handler) ListTenants(c echo.Context, successMessage, errMessage string, confirmDelete bool) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	// if we confirm an action let's save the tenantID
	if confirmDelete {
		commonInfo.ActionTenantID = commonInfo.TenantID
	}
	// Override tenant and site ids as we're working in global config
	commonInfo.TenantID = "-1"
	commonInfo.SiteID = "-1"

	f := filters.TenantFilter{}

	nameFilter := c.FormValue("filterByName")
	if nameFilter != "" {
		f.Name = nameFilter
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

	p.NItems, err = h.Model.CountAllTenants(f)
	if err != nil {
		successMessage = ""
		errMessage = err.Error()
	}

	tenants, err := h.Model.GetTenantsByPage(p, f)
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

	return RenderView(c, admin_views.TenantsIndex(" | Tenants", admin_views.Tenants(c, p, f, tenants, successMessage, errMessage, refreshTime, agentsExists, serversExists, confirmDelete, commonInfo), commonInfo))
}

func (h *Handler) NewTenant(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
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

	return RenderView(c, admin_views.TenantsIndex(" | Tenants", admin_views.NewTenant(c, defaultCountry, agentsExists, serversExists, commonInfo), commonInfo))
}

func (h *Handler) AddTenant(c echo.Context) error {
	successMessage := ""
	errMessage := ""

	name := c.FormValue("name")
	if name == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.name_cannot_be_empty"), true))
	}

	exists, err := h.Model.TenantNameTaken(name)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_check_name", err.Error()), true))
	}

	if exists {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.tenant_name_taken", name), true))
	}

	isDefault, err := strconv.ParseBool(c.FormValue("is-default"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_bool", err.Error()), true))
	}

	siteName := c.FormValue("site-name")
	if siteName == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.site_name_cannot_be_empty"), true))
	}

	err = h.Model.AddTenant(name, isDefault, siteName)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.new_error"), true))
	}

	successMessage = i18n.T(c.Request().Context(), "tenants.new_success")
	return h.ListTenants(c, successMessage, errMessage, false)
}

func (h *Handler) EditTenant(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_get_common_info", err.Error()), true))
	}

	// Override tenant and site ids as we're working in global config
	commonInfo.TenantID = "-1"
	commonInfo.SiteID = "-1"

	id := c.Param("tenant")
	if id == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.tenant_cannot_be_empty"), true))
	}

	tenantID, err := strconv.Atoi(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", id), true))
	}

	t, err := h.Model.GetTenantByID(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.tenant_not_found", err.Error()), true))
	}

	if c.Request().Method == "POST" {
		name := c.FormValue("name")
		if name == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.name_cannot_be_empty"), true))
		}

		if t.Description != name {
			exists, err := h.Model.TenantNameTaken(name)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_check_name", err.Error()), true))
			}

			if exists {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.tenant_name_taken", name), true))
			}
		}

		isDefault, err := strconv.ParseBool(c.FormValue("is-default"))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_bool", err.Error()), true))
		}

		if err := h.Model.UpdateTenant(t.ID, name, isDefault); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}

		return h.ListTenants(c, i18n.T(c.Request().Context(), "tenants.edit_success"), "", false)
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

	return RenderView(c, admin_views.TenantsIndex(" | Tenants", admin_views.EditTenant(c, t, defaultCountry, agentsExists, serversExists, commonInfo), commonInfo))
}

func (h *Handler) DeleteTenant(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.could_not_get_common_info", err.Error()), false)
	}

	// Override tenant and site ids as we're working in global config
	commonInfo.TenantID = "-1"
	commonInfo.SiteID = "-1"

	id := c.Param("tenant")
	if id == "" {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.tenant_cannot_be_empty"), false)
	}

	tenantID, err := strconv.Atoi(id)
	if err != nil {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", id), false)
	}

	nTenants, err := h.Model.CountTenants()
	if err != nil {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.could_not_count_tenants"), false)
	}

	if nTenants == 1 {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.at_least_one_tenant"), false)
	}

	t, err := h.Model.GetTenantByID(tenantID)
	if err != nil {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.could_not_find_tenant"), false)
	}
	if t.IsDefault {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.default_cannot_be_deleted"), false)
	}

	// Send a request to uninstall agents associated with this organization
	agents, err := h.Model.GetAgentsByTenant(tenantID)
	if err != nil {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.could_not_get_agents"), false)
	}

	if h.NATSConnection == nil || !h.NATSConnection.IsConnected() {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "nats.not_connected"), false)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, a := range agents {
		if _, err := h.JetStream.Publish(ctx, "agent.uninstall."+a.ID, nil); err != nil {
			return h.ListTenants(c, "", i18n.T(c.Request().Context(), "agents.could_not_send_request_to_uninstall"), false)
		}
	}

	// Remove the tenant with cascade
	if err := h.Model.DeleteTenant(tenantID); err != nil {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.delete_error", err.Error()), false)
	}

	successMessage := i18n.T(c.Request().Context(), "tenants.deleted")
	return h.ListTenants(c, successMessage, "", false)
}

func (h *Handler) ImportTenants(c echo.Context) error {
	var errors = []string{}

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
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.import_read_error", err.Error()), false))

		}

		if len(record) != 2 {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.import_error_wrong_format", index), false))
		}

		if record[0] == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.import_required", "orgname", index), false))
		}

		if record[1] == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.import_required", "sitename", index), false))
		}

		index++

		exists, err := h.Model.TenantNameTaken(record[0])
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		if exists {
			errors = append(errors, i18n.T(c.Request().Context(), "tenants.tenant_name_taken", record[0]))
			continue
		}

		err = h.Model.AddTenant(record[0], false, record[1])
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
	}

	if len(errors) > 0 {
		return h.ListTenants(c, "", i18n.T(c.Request().Context(), "tenants.import_wrong_tenants", strings.Join(errors, ",")), false)
	}

	return h.ListTenants(c, i18n.T(c.Request().Context(), "tenants.import_success"), "", false)
}
