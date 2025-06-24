package handlers

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/tasks_views"
	"github.com/open-uem/wingetcfg/wingetcfg"
)

func (h *Handler) NewTask(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
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

	return RenderView(c, tasks_views.TasksIndex("| Tasks", tasks_views.NewTask(c, profileID, commonInfo), commonInfo))
}

func (h *Handler) EditTask(c echo.Context) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
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

	return RenderView(c, tasks_views.TasksIndex("| Tasks", tasks_views.EditTask(c, task.Edges.Profile.ID, task, commonInfo), commonInfo))
}

func validateTaskForm(c echo.Context) (*models.TaskConfig, error) {
	taskConfig := models.TaskConfig{}

	validTasks := []string{
		"winget_install",
		"winget_delete",
		"add_registry_key",
		"remove_registry_key",
		"update_registry_key_default_value",
		"add_registry_key_value",
		"remove_registry_key_value",
		"add_local_user",
		"remove_local_user",
		"add_local_group",
		"remove_local_group",
		"add_users_to_local_group",
		"remove_users_from_local_group",
		"msi_install",
		"msi_uninstall",
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	if c.FormValue("package-task-type") != "" {
		taskConfig.TaskType = c.FormValue("package-task-type")
	}
	if c.FormValue("registry-task-type") != "" {
		taskConfig.TaskType = c.FormValue("registry-task-type")
	}
	if c.FormValue("local-user-task-type") != "" {
		taskConfig.TaskType = c.FormValue("local-user-task-type")
	}
	if c.FormValue("local-group-task-type") != "" {
		taskConfig.TaskType = c.FormValue("local-group-task-type")
	}
	if c.FormValue("msi-task-type") != "" {
		taskConfig.TaskType = c.FormValue("msi-task-type")
	}
	if c.FormValue("selected-task-type") != "" {
		taskConfig.TaskType = c.FormValue("selected-task-type")
	}

	if !slices.Contains(validTasks, taskConfig.TaskType) {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.wrong_type"))
	}

	taskConfig.ExecuteCommand = c.FormValue("execute-command")
	if taskConfig.TaskType == "execute_command" && taskConfig.ExecuteCommand == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.execute_command_not_empty"))
	}

	// Package management

	taskConfig.PackageID = c.FormValue("package-id")
	if (taskConfig.TaskType == "winget_install" || taskConfig.TaskType == "winget_delete") && taskConfig.PackageID == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.package_id_not_empty"))
	}

	taskConfig.PackageName = c.FormValue("package-name")
	if (taskConfig.TaskType == "winget_install" || taskConfig.TaskType == "winget_delete") && taskConfig.PackageName == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.package_name_not_empty"))
	}

	// Registry management

	taskConfig.RegistryKey = c.FormValue("registry-key")
	if (taskConfig.TaskType == "add_registry_key" || taskConfig.TaskType == "remove_registry_key") && taskConfig.RegistryKey == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.registry_key_not_empty"))
	}

	taskConfig.RegistryKeyValue = c.FormValue("registry-value-name")
	if (taskConfig.TaskType == "add_registry_key_value" || taskConfig.TaskType == "remove_registry_key_value") && taskConfig.RegistryKeyValue == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.invalid_registry_value_name"))
	}

	taskConfig.RegistryKeyValueType = c.FormValue("registry-value-type")
	if !slices.Contains([]string{"", wingetcfg.RegistryValueTypeString, wingetcfg.RegistryValueTypeDWord, wingetcfg.RegistryValueTypeQWord, wingetcfg.RegistryValueTypeMultistring}, taskConfig.RegistryKeyValueType) {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.invalid_registry_value_type"))
	}

	taskConfig.RegistryKeyValueData = c.FormValue("registry-value-data")
	if (taskConfig.TaskType == "update_registry_key_default_value") && taskConfig.RegistryKeyValueData == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.invalid_registry_value_data"))
	}

	dataStrings := strings.Split(taskConfig.RegistryKeyValueData, "\n")
	if len(dataStrings) > 1 && taskConfig.RegistryKeyValueType != wingetcfg.RegistryValueTypeMultistring {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.unexpected_multiple_strings"))
	}

	registryKeyHex := c.FormValue("registry-hex")
	if registryKeyHex == "on" {
		taskConfig.RegistryHex = true
	}
	if taskConfig.RegistryKeyValueType != wingetcfg.RegistryValueTypeDWord && taskConfig.RegistryKeyValueType != wingetcfg.RegistryValueTypeQWord && taskConfig.RegistryHex {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.unexpected_hex"))
	}

	registryKeyForce := c.FormValue("registry-key-force")
	registryValueForce := c.FormValue("registry-value-force")

	if registryKeyForce == "on" || registryValueForce == "on" {
		taskConfig.RegistryForce = true
	}

	// Local User
	taskConfig.LocalUserUsername = c.FormValue("local-user-username")
	if (taskConfig.TaskType == "add_local_user" || taskConfig.TaskType == "remove_local_user") && taskConfig.LocalUserUsername == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_username_is_required"))
	}

	taskConfig.LocalUserDescription = c.FormValue("local-user-description")
	taskConfig.LocalUserFullName = c.FormValue("local-user-fullname")
	taskConfig.LocalUserPassword = c.FormValue("local-user-password")

	localUserDisabled := c.FormValue("local-user-disabled")
	if localUserDisabled == "on" {
		taskConfig.LocalUserDisabled = true
	}

	localUserPasswordChangeNotAllowed := c.FormValue("local-user-password-change-disallow")
	if localUserPasswordChangeNotAllowed == "on" {
		taskConfig.LocalUserPasswordChangeNotAllowed = true
	}

	localUserPasswordChangeRequired := c.FormValue("local-user-password-change-required")
	if localUserPasswordChangeRequired == "on" {
		taskConfig.LocalUserPasswordChangeRequired = true
	}

	localUserNeverExpires := c.FormValue("local-user-password-never-expires")
	if localUserNeverExpires == "on" {
		taskConfig.LocalUserNeverExpires = true
	}

	// Local group
	taskConfig.LocalGroupName = c.FormValue("local-group-name")
	if (taskConfig.TaskType == "add_local_group" || taskConfig.TaskType == "remove_local_group" || taskConfig.TaskType == "add_users_to_local_group" || taskConfig.TaskType == "remove_users_from_local_group") && taskConfig.LocalGroupName == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_group_name_is_required"))
	}

	taskConfig.LocalGroupDescription = c.FormValue("local-group-description")
	if taskConfig.TaskType == "add_local_group" && taskConfig.LocalGroupName == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_group_description_is_required"))
	}

	taskConfig.LocalGroupMembers = c.FormValue("local-group-members")

	taskConfig.LocalGroupMembersToInclude = c.FormValue("local-group-members-to-include")
	if taskConfig.LocalGroupMembersToInclude != "" && taskConfig.LocalGroupMembers != "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_group_members_included_and_members_exclusive"))
	}

	taskConfig.LocalGroupMembersToExclude = c.FormValue("local-group-members-to-exclude")
	if taskConfig.LocalGroupMembersToExclude != "" && taskConfig.LocalGroupMembers != "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_group_members_excluded_and_members_exclusive"))
	}

	// MSI
	taskConfig.MsiProductID = c.FormValue("msi-productid")
	if (taskConfig.TaskType == "msi_install" || taskConfig.TaskType == "msi_uninstall") && taskConfig.MsiProductID == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.msi_productid_not_empty"))
	}

	taskConfig.MsiPath = c.FormValue("msi-path")
	if (taskConfig.TaskType == "msi_install" || taskConfig.TaskType == "msi_uninstall") && taskConfig.MsiPath == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.msi_path_not_empty"))
	}

	taskConfig.MsiArguments = c.FormValue("msi-arguments")
	taskConfig.MsiLogPath = c.FormValue("msi-log-path")
	taskConfig.MsiFileHash = c.FormValue("msi-hash")

	if taskConfig.MsiHashAlgorithm != "" &&
		taskConfig.MsiHashAlgorithm != wingetcfg.FileHashMD5 &&
		taskConfig.MsiHashAlgorithm != wingetcfg.FileHashRIPEMD160 &&
		taskConfig.MsiHashAlgorithm != wingetcfg.FileHashSHA1 &&
		taskConfig.MsiHashAlgorithm != wingetcfg.FileHashSHA256 &&
		taskConfig.MsiHashAlgorithm != wingetcfg.FileHashSHA384 &&
		taskConfig.MsiHashAlgorithm != wingetcfg.FileHashSHA512 {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.unexpected_msi_hash_algorithm"))
	}
	taskConfig.MsiHashAlgorithm = c.FormValue("msi-hash-alg")

	if (taskConfig.TaskType == "msi_install" || taskConfig.TaskType == "msi_uninstall") &&
		((taskConfig.MsiFileHash == "" && taskConfig.MsiHashAlgorithm != "") || (taskConfig.MsiFileHash != "" && taskConfig.MsiHashAlgorithm == "")) {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.msi_specify_both_hash_inputs"))
	}

	return &taskConfig, nil
}
