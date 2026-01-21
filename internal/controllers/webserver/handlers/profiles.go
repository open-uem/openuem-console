package handlers

import (
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/ent/task"
	"github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/profiles_views"
)

func (h *Handler) Profiles(c echo.Context, successMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	p.NItems, err = h.Model.CountAllProfiles(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	profiles, err := h.Model.GetProfilesByPage(p, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	confirmDelete := false
	profileId := ""

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.Profiles(c, p, profiles, refreshTime, profileId, confirmDelete, successMessage, itemsPerPage, commonInfo), commonInfo))
}

func (h *Handler) NewProfile(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	siteID, err := strconv.Atoi(commonInfo.SiteID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "sites.could_not_convert_site_to_int", commonInfo.SiteID), true))
	}

	if siteID == -1 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.profile_empty_site", commonInfo.SiteID), true))
	}

	if c.Request().Method == "POST" {
		description := c.FormValue("profile-description")
		if description == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.new.empty"), true))
		}

		profile, err := h.Model.AddProfile(siteID, description)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.new.could_not_save"), true))
		}

		return h.EditProfile(c, "GET", strconv.Itoa(profile.ID), i18n.T(c.Request().Context(), "profiles.new.saved"))
	}

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.NewProfile(c, commonInfo), commonInfo))
}

func (h *Handler) EditProfile(c echo.Context, method string, id string, successMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	if id == "" {
		id = c.Param("uuid")
		if id == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.empty_id"), true))
		}
	}

	profileId, err := strconv.Atoi(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.invalid_task"), true))
	}

	profile, err := h.Model.GetProfileById(profileId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_err"), true))
	}

	p.NItems, err = h.Model.CountAllTasksForProfile(profileId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_tasks_err"), true))
	}

	tasks, err := h.Model.GetTasksForProfileByPage(p, profileId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_tasks_err"), true))
	}

	tags, err := h.Model.GetAllTags(commonInfo, filters.AgentFilter{})
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.no_tags"), true))
	}

	if method == "" {
		method = c.Request().Method
	}

	if method == "POST" {
		description := c.FormValue("profile-description")
		if description == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.empty"), true))
		}

		applyToAll := c.FormValue("profile-assignment")

		if err := h.Model.UpdateProfile(profileId, description, applyToAll, commonInfo); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.could_not_save"), true))
		}

		return h.EditProfile(c, "GET", id, i18n.T(c.Request().Context(), "profiles.edit.saved"))
	}

	if method == "DELETE" {
		if err := h.Model.DeleteProfile(profileId, commonInfo); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.could_not_delete"), true))
		}
		return h.Profiles(c, i18n.T(c.Request().Context(), "profiles.edit.deleted"))
	}

	confirmDelete := false

	if successMessage != "" {
		u, err := url.Parse(partials.GetNavigationUrl(commonInfo, fmt.Sprintf("/profiles/%s", id)))
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}
		return RenderViewWithReplaceUrl(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.EditProfile(c, p, profile, tasks, tags, "", successMessage, confirmDelete, itemsPerPage, commonInfo), commonInfo), u)
	}

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.EditProfile(c, p, profile, tasks, tags, "", successMessage, confirmDelete, itemsPerPage, commonInfo), commonInfo))
}

func (h *Handler) ProfileTags(c echo.Context) error {
	id := c.Param("uuid")
	if id == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.empty_id"), true))
	}

	tag := c.FormValue("tagId")
	if tag == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.empty_tag_id"), true))
	}

	profileId, err := strconv.Atoi(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.invalid_task"), true))
	}

	tagId, err := strconv.Atoi(tag)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.tag_id_invalid"), true))
	}

	if c.Request().Method == "POST" {
		if err := h.Model.AddTagToProfile(profileId, tagId); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.could_not_add_tag"), true))
		}
		return h.EditProfile(c, "GET", id, i18n.T(c.Request().Context(), "profiles.edit.tag_added"))
	}

	if c.Request().Method == "DELETE" {
		if err := h.Model.RemoveTagFromProfile(profileId, tagId); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.could_not_remove_tag"), true))
		}
		return h.EditProfile(c, "GET", id, i18n.T(c.Request().Context(), "profiles.edit.tag_removed"))
	}

	return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.wrong_method"), true))
}

func (h *Handler) ConfirmDeleteProfile(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	profileId := c.Param("uuid")
	if profileId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.empty_id"), true))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	p.NItems, err = h.Model.CountAllProfiles(commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	profiles, err := h.Model.GetProfilesByPage(p, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	confirmDelete := true
	successMessage := ""

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.Profiles(c, p, profiles, refreshTime, profileId, confirmDelete, successMessage, itemsPerPage, commonInfo), commonInfo))
}

func (h *Handler) ConfirmDeleteTask(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	id := c.Param("profile")
	if id == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.empty_id"), true))
	}

	profileId, err := strconv.Atoi(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.invalid_task"), true))
	}

	profile, err := h.Model.GetProfileById(profileId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_err"), true))
	}

	p.NItems, err = h.Model.CountAllTasksForProfile(profileId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_tasks_err"), true))
	}

	tasks, err := h.Model.GetTasksForProfileByPage(p, profileId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_tasks_err"), true))
	}

	tags, err := h.Model.GetAllTags(commonInfo, filters.AgentFilter{})
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.no_tags"), true))
	}

	taskId := c.Param("task")
	if taskId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.edit.empty_task"), true))
	}

	taskIdAsInt, err := strconv.Atoi(taskId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.edit.invalid_task"), true))
	}

	_, err = h.Model.GetTasksById(taskIdAsInt)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.edit.could_not_get"), true))
	}

	successMessage := ""
	confirmDelete := true

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.EditProfile(c, p, profile, tasks, tags, taskId, successMessage, confirmDelete, itemsPerPage, commonInfo), commonInfo))
}

func (h *Handler) ProfileIssues(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	profileID := c.Param("uuid")
	if profileID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.issues.empty_id"), true))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.NewPaginationAndSort(itemsPerPage)
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"), itemsPerPage)

	pID, err := strconv.Atoi(profileID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p.NItems, err = h.Model.CountAllProfileIssues(pID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	profile, err := h.Model.GetProfileById(pID, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_err"), true))
	}

	issues, err := h.Model.GetProfileIssuesByPage(p, pID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.ProfilesIssues(c, p, issues, profile, itemsPerPage, commonInfo), commonInfo))
}

func (h *Handler) ProfileTaskTypes(c echo.Context) error {
	agentType := c.QueryParam("task-agent-type")
	if agentType != "" {
		return RenderView(c, partials.SelectTaskType(nil, agentType))
	}
	return nil
}

func (h *Handler) ProfileTaskSubTypes(c echo.Context) error {
	taskType := c.QueryParam("task-type")

	switch taskType {
	case "package_type":
		return RenderView(c, partials.SelectWinGetPackageTaskSubtype(nil))
	case "registry_type":
		return RenderView(c, partials.SelectRegistryTaskSubtype(nil))
	case "local_user_subtype":
		return RenderView(c, partials.SelectLocalUserTaskSubtype(nil))
	case "local_group_subtype":
		return RenderView(c, partials.SelectWindowsLocalGroupTaskSubtype(nil))
	case "unix_local_user_subtype":
		return RenderView(c, partials.SelectUnixLocalUserTaskSubtype(nil))
	case "unix_local_group_subtype":
		return RenderView(c, partials.SelectUnixLocalGroupTaskSubtype(nil))
	case "msi_type":
		return RenderView(c, partials.SelectMSITaskSubtype(nil))
	case "powershell_type":
		c.Response().Header().Set("HX-Retarget", "#task-definition")
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
		c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")
		c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		return partials.PowerShellComponent(nil).Render(c.Request().Context(), c.Response().Writer)
	case "unix_script_type":
		c.Response().Header().Set("HX-Retarget", "#task-definition")
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
		c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")
		c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		return partials.UnixScriptComponent(nil).Render(c.Request().Context(), c.Response().Writer)
	case "flatpak_type":
		return RenderView(c, partials.SelectFlatpakPackageTaskSubtype(nil))
	case "brew_formula_type":
		return RenderView(c, partials.SelectHomeBrewFormulaTaskSubtype(nil))
	case "brew_cask_type":
		return RenderView(c, partials.SelectHomeBrewCaskTaskSubtype(nil))
	case "netbird_type":
		return RenderView(c, partials.SelectNetbirdTaskSubtype(nil))
		// Future APT
		// case "apt_type":
		// 	return RenderView(c, partials.SelectAPTTaskSubtype(nil))
	}

	return nil
}

func (h *Handler) ProfileTaskDefinition(c echo.Context) error {
	taskType := c.QueryParam("task-subtype")

	t := ent.Task{}
	switch taskType {
	case task.TypeWingetInstall.String(), task.TypeWingetDelete.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.WingetPackageSearch(&t))
	case task.TypeAddRegistryKey.String(), task.TypeAddRegistryKeyValue.String(), task.TypeUpdateRegistryKeyDefaultValue.String(),
		task.TypeRemoveRegistryKey.String(), task.TypeRemoveRegistryKeyValue.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.RegistryComponent(&t))
	case task.TypeAddLocalUser.String(), task.TypeRemoveLocalUser.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.LocalUserComponent(&t))
	case task.TypeAddLocalGroup.String(), task.TypeRemoveLocalGroup.String(), task.TypeAddUsersToLocalGroup.String(), task.TypeRemoveUsersFromLocalGroup.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.WindowsLocalGroupComponent(&t))
	case task.TypeAddUnixLocalGroup.String(), task.TypeRemoveUnixLocalGroup.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.UnixLocalGroupComponent(&t))
	case task.TypeAddUnixLocalUser.String(), task.TypeModifyUnixLocalUser.String(), task.TypeRemoveUnixLocalUser.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.UnixLocalUserComponent(&t))
	case task.TypeMsiInstall.String(), task.TypeMsiUninstall.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.MSIComponent(&t))
	case task.TypeFlatpakInstall.String(), task.TypeFlatpakUninstall.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.FlatpakPackageManagement(&t))
	case task.TypeBrewFormulaInstall.String(), task.TypeBrewFormulaUninstall.String(), task.TypeBrewFormulaUpgrade.String(), task.TypeBrewCaskInstall.String(), task.TypeBrewCaskUninstall.String(), task.TypeBrewCaskUpgrade.String():
		t.Type = task.Type(taskType)
		return RenderView(c, partials.HomeBrewPackageManagement(&t))
	case task.TypeNetbirdInstall.String(), task.TypeNetbirdUninstall.String(), task.TypeNetbirdRegister.String():
		t.Type = task.Type(taskType)

		ng := []nats.NetBirdGroups{}
		if taskType == task.TypeNetbirdRegister.String() {

			commonInfo, err := h.GetCommonInfo(c)
			if err != nil {
				return err
			}

			tenantID, err := strconv.Atoi(commonInfo.TenantID)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
			}

			settings, err := h.Model.GetNetbirdSettings(tenantID)
			if err != nil {
				if ent.IsNotFound(err) {
					settings = &ent.NetbirdSettings{}
				} else {
					return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
				}
			}

			if settings.AccessToken == "" {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.token_empty"), true))
			}

			ng, err = getGroupsFromNetbirdAPI(settings.ManagementURL, settings.AccessToken)
			if err != nil {
				return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_groups", err.Error()), true))
			}
		}

		return RenderView(c, partials.NetbirdTaskComponent(&t, ng))
	}

	return nil
}
