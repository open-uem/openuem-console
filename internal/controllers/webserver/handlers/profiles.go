package handlers

import (
	"log"
	"net/url"
	"strconv"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/profiles_views"
)

func (h *Handler) Profiles(c echo.Context, successMessage string) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	p.NItems, err = h.Model.CountAllProfiles()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	profiles, err := h.Model.GetProfilesByPage(p)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	l := views.GetTranslatorForDates(c)

	confirmDelete := false
	profileId := ""

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.Profiles(c, p, l, h.SessionManager, profiles, refreshTime, h.Version, latestServerRelease.Version, profileId, confirmDelete, successMessage)))
}

func (h *Handler) NewProfile(c echo.Context) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	if c.Request().Method == "POST" {
		description := c.FormValue("profile-description")
		if description == "" {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.new.empty"), true))
		}

		profile, err := h.Model.AddProfile(description)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.new.could_not_save"), true))
		}

		return h.EditProfile(c, "GET", strconv.Itoa(profile.ID), i18n.T(c.Request().Context(), "profiles.new.saved"))
	}

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.NewProfile(h.SessionManager, h.Version, latestServerRelease.Version)))
}

func (h *Handler) EditProfile(c echo.Context, method string, id string, successMessage string) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

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

	profile, err := h.Model.GetProfileById(profileId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_err"), true))
	}

	p.NItems, err = h.Model.CountAllTasksForProfile(profileId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_tasks_err"), true))
	}

	tasks, err := h.Model.GetTasksForProfileByPage(p, profileId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_tasks_err"), true))
	}

	tags, err := h.Model.GetAllTags()
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

		if err := h.Model.UpdateProfile(profileId, description, applyToAll); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.could_not_save"), true))
		}

		return h.EditProfile(c, "GET", id, i18n.T(c.Request().Context(), "profiles.edit.saved"))
	}

	if method == "DELETE" {
		if err := h.Model.DeleteProfile(profileId); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.could_not_delete"), true))
		}
		return h.Profiles(c, i18n.T(c.Request().Context(), "profiles.edit.deleted"))
	}

	confirmDelete := false

	if successMessage != "" {
		u, err := url.Parse("/profiles/" + id)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), true))
		}
		return RenderViewWithReplaceUrl(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.EditProfile(c, p, h.SessionManager, profile, tasks, tags, "", h.Version, latestServerRelease.Version, successMessage, confirmDelete)), u)
	}

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.EditProfile(c, p, h.SessionManager, profile, tasks, tags, "", h.Version, latestServerRelease.Version, successMessage, confirmDelete)))
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
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	profileId := c.Param("uuid")
	if profileId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.empty_id"), true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	p.NItems, err = h.Model.CountAllProfiles()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	profiles, err := h.Model.GetProfilesByPage(p)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	refreshTime, err := h.Model.GetDefaultRefreshTime()
	if err != nil {
		log.Println("[ERROR]: could not get refresh time from database")
		refreshTime = 5
	}

	l := views.GetTranslatorForDates(c)
	confirmDelete := true
	successMessage := ""

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.Profiles(c, p, l, h.SessionManager, profiles, refreshTime, h.Version, latestServerRelease.Version, profileId, confirmDelete, successMessage)))
}

func (h *Handler) ConfirmDeleteTask(c echo.Context) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	id := c.Param("profile")
	if id == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.empty_id"), true))
	}

	profileId, err := strconv.Atoi(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.invalid_task"), true))
	}

	profile, err := h.Model.GetProfileById(profileId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_err"), true))
	}

	p.NItems, err = h.Model.CountAllTasksForProfile(profileId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_tasks_err"), true))
	}

	tasks, err := h.Model.GetTasksForProfileByPage(p, profileId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_tasks_err"), true))
	}

	tags, err := h.Model.GetAllTags()
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
	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.EditProfile(c, p, h.SessionManager, profile, tasks, tags, taskId, h.Version, latestServerRelease.Version, successMessage, confirmDelete)))
}

func (h *Handler) ProfileIssues(c echo.Context) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	profileID := c.Param("uuid")
	if profileID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.issues.empty_id"), true))
	}

	p := partials.NewPaginationAndSort()
	p.GetPaginationAndSortParams(c.FormValue("page"), c.FormValue("pageSize"), c.FormValue("sortBy"), c.FormValue("sortOrder"), c.FormValue("currentSortBy"))

	pID, err := strconv.Atoi(profileID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	p.NItems, err = h.Model.CountAllProfileIssues(pID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	profile, err := h.Model.GetProfileById(pID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "profiles.edit.retrieve_err"), true))
	}

	issues, err := h.Model.GetProfileIssuesByPage(p, pID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	l := views.GetTranslatorForDates(c)

	return RenderView(c, profiles_views.ProfilesIndex("| Profiles", profiles_views.ProfilesIssues(c, p, l, h.SessionManager, issues, profile, h.Version, latestServerRelease.Version)))
}
