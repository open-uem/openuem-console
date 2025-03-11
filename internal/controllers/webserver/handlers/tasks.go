package handlers

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/models"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/tasks_views"
	"github.com/open-uem/wingetcfg/wingetcfg"
)

func (h *Handler) NewTask(c echo.Context) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	profile := c.Param("profile")
	if profile == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.new.empty_profile"), true))
	}

	profileID, err := strconv.Atoi(profile)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.new.invalid_profile"), true))
	}

	if c.Request().Method == "POST" {
		t, err := validateTaskForm(c)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%v", err), true))
		}

		if err := h.Model.AddTaskToProfile(c, profileID, *t); err != nil {
			return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%s : %v", i18n.T(c.Request().Context(), "tasks.new.could_not_save"), err), true))
		}

		return h.EditProfile(c, "GET", profile, i18n.T(c.Request().Context(), "tasks.new.saved"))
	}

	return RenderView(c, tasks_views.TasksIndex("| Tasks", tasks_views.NewTask(c, h.SessionManager, profileID, h.Version, latestServerRelease.Version)))
}

func (h *Handler) EditTask(c echo.Context) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	id := c.Param("id")
	if id == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.edit.empty_task"), true))
	}

	taskId, err := strconv.Atoi(id)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.edit.invalid_task"), true))
	}

	task, err := h.Model.GetTasksById(taskId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%s : %v", i18n.T(c.Request().Context(), "tasks.edit.could_not_save"), err), true))
	}

	if task.Edges.Profile == nil {
		return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%s : %v", i18n.T(c.Request().Context(), "tasks.edit.no_profile"), err), true))
	}

	if c.Request().Method == "POST" {
		t, err := validateTaskForm(c)
		if err != nil {
			return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%v", err), true))
		}

		if err := h.Model.UpdateTaskToProfile(c, taskId, *t); err != nil {
			return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%s : %v", i18n.T(c.Request().Context(), "tasks.edit.could_not_save"), err), true))
		}

		return h.EditProfile(c, "GET", strconv.Itoa(task.Edges.Profile.ID), i18n.T(c.Request().Context(), "tasks.edit.saved"))
	}

	if c.Request().Method == "DELETE" {
		if err := h.Model.DeleteTask(taskId); err != nil {
			return RenderError(c, partials.ErrorMessage(fmt.Sprintf("%s : %v", i18n.T(c.Request().Context(), "tasks.edit.could_not_delete"), err), true))
		}
		return h.EditProfile(c, "GET", strconv.Itoa(task.Edges.Profile.ID), i18n.T(c.Request().Context(), "tasks.edit.deleted"))
	}

	return RenderView(c, tasks_views.TasksIndex("| Tasks", tasks_views.EditTask(c, h.SessionManager, task.Edges.Profile.ID, task, h.Version, latestServerRelease.Version)))
}

func validateTaskForm(c echo.Context) (*models.TaskConfig, error) {
	taskConfig := models.TaskConfig{}

	validTasks := []string{"winget_install", "winget_delete", "add_registry_key", "remove_registry_key", "update_registry_key_default_value", "add_registry_key_value", "remove_registry_key_value"}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	if c.FormValue("package-task-type") != "" {
		taskConfig.TaskType = c.FormValue("package-task-type")
	}
	if c.FormValue("registry-task-type") != "" {
		taskConfig.TaskType = c.FormValue("registry-task-type")
	}
	if c.FormValue("selected-task-type") != "" {
		taskConfig.TaskType = c.FormValue("selected-task-type")
	}

	if !slices.Contains(validTasks, taskConfig.TaskType) {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.new.wrong_type"))
	}

	taskConfig.ExecuteCommand = c.FormValue("execute-command")
	if taskConfig.TaskType == "execute_command" && taskConfig.ExecuteCommand == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.execute_command_not_empty"))
	}

	taskConfig.PackageID = c.FormValue("package-id")
	if (taskConfig.TaskType == "winget_install" || taskConfig.TaskType == "winget_delete") && taskConfig.PackageID == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.package_id_not_empty"))
	}

	taskConfig.PackageName = c.FormValue("package-name")
	if (taskConfig.TaskType == "winget_install" || taskConfig.TaskType == "winget_delete") && taskConfig.PackageName == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.package_name_not_empty"))
	}

	taskConfig.RegistryKey = c.FormValue("registry-key")
	if (taskConfig.TaskType == "add_registry_key" || taskConfig.TaskType == "remove_registry_key") && taskConfig.RegistryKey == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.registry_key_not_empty"))
	}

	taskConfig.RegistryKeyValue = c.FormValue("registry-value-name")
	if (taskConfig.TaskType == "add_registry_key_value" || taskConfig.TaskType == "remove_registry_key_value") && taskConfig.RegistryKeyValue == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.invalid_registry_value_name"))
	}

	taskConfig.RegistryKeyValueType = c.FormValue("registry-value-type")
	if !slices.Contains([]string{"", wingetcfg.RegistryValueTypeString, wingetcfg.RegistryValueTypeDWord, wingetcfg.RegistryValueTypeQWord, wingetcfg.RegistryValueTypeMultistring}, taskConfig.RegistryKeyValueType) {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.invalid_registry_value_type"))
	}

	taskConfig.RegistryKeyValueData = c.FormValue("registry-value-data")
	if (taskConfig.TaskType == "update_registry_key_default_value") && taskConfig.RegistryKeyValueData == "" {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.invalid_registry_value_data"))
	}

	dataStrings := strings.Split(taskConfig.RegistryKeyValueData, "\n")
	if len(dataStrings) > 1 && taskConfig.RegistryKeyValueType != wingetcfg.RegistryValueTypeMultistring {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.unexpected_multiple_strings"))
	}

	registryKeyHex := c.FormValue("registry-hex")
	if registryKeyHex == "on" {
		taskConfig.RegistryHex = true
	}
	if taskConfig.RegistryKeyValueType != wingetcfg.RegistryValueTypeDWord && taskConfig.RegistryKeyValueType != wingetcfg.RegistryValueTypeQWord && taskConfig.RegistryHex {
		return nil, fmt.Errorf(i18n.T(c.Request().Context(), "tasks.unexpected_hex"))
	}

	registryKeyForce := c.FormValue("registry-key-force")
	registryValueForce := c.FormValue("registry-value-force")

	if registryKeyForce == "on" || registryValueForce == "on" {
		taskConfig.RegistryForce = true
	}

	return &taskConfig, nil
}
