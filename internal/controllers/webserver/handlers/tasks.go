package handlers

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent/task"
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
	taskType := ""

	if c.FormValue("task-subtype") != "" {
		taskType = c.FormValue("task-subtype")
	}

	if c.FormValue("powershell-script") != "" {
		taskType = "powershell_script"
	}

	if c.FormValue("unix-script") != "" {
		taskType = "unix_script"
	}

	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskID := c.Param("id")
	agentsType := c.FormValue("task-agent-type")
	if taskID == "" && (agentsType == "" || !slices.Contains([]string{"windows", "linux", "macos"}, agentsType)) {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.wrong_agenttype"))
	}

	switch taskType {
	case string(task.TypeAddLocalUser), string(task.TypeRemoveLocalUser):
		return validateWindowsLocalUser(c)
	case string(task.TypeAddRegistryKey), string(task.TypeAddRegistryKeyValue), string(task.TypeRemoveRegistryKey),
		string(task.TypeRemoveRegistryKeyValue), string(task.TypeUpdateRegistryKeyDefaultValue):
		return validateWindowsRegistry(c)
	case string(task.TypeAddLinuxLocalUser):
		return validateAddLinuxLocalUser(c)
	case string(task.TypeRemoveLinuxLocalUser):
		return validateRemoveLinuxLocalUser(c)
	case string(task.TypeAddLocalGroup), string(task.TypeRemoveLocalGroup):
		return validateWindowsLocalGroup(c)
	case string(task.TypeAddUnixLocalGroup), string(task.TypeRemoveUnixLocalGroup):
		return validateUnixLocalGroup(c)
	case string(task.TypeMsiInstall), string(task.TypeMsiUninstall):
		return validateMSI(c)
	case string(task.TypeWingetDelete), string(task.TypeWingetInstall), string(task.TypeWingetUpdate):
		return validateWinGetPackage(c)
	case string(task.TypePowershellScript):
		return validatePowerShellScript(c)
	case string(task.TypeUnixScript):
		return validateUnixScript(c)
	case string(task.TypeFlatpakInstall), string(task.TypeFlatpakUninstall):
		return validateFlatpakWinGetPackage(c)
	default:
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.wrong_type"))
	}
}

func validateWindowsRegistry(c echo.Context) (*models.TaskConfig, error) {
	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

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

	return &taskConfig, nil
}

func validateWindowsLocalUser(c echo.Context) (*models.TaskConfig, error) {
	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	taskConfig.LocalUserUsername = c.FormValue("local-user-username")
	if taskConfig.LocalUserUsername == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_username_is_required"))
	}

	taskConfig.LocalUserDescription = c.FormValue("local-user-description")
	taskConfig.LocalUserFullName = c.FormValue("local-user-fullname")
	password := c.FormValue("local-user-password")
	if password != "" {
		taskConfig.LocalUserPassword = c.FormValue("local-user-password")
	}

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

	return &taskConfig, nil
}

func validateAddLinuxLocalUser(c echo.Context) (*models.TaskConfig, error) {
	var err error

	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	taskConfig.LocalUserUsername = c.FormValue("local-user-username")
	if taskConfig.LocalUserUsername == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_username_is_required"))
	}

	taskConfig.LocalUserDescription = c.FormValue("local-user-description")
	taskConfig.LocalUserFullName = c.FormValue("local-user-fullname")
	password := c.FormValue("local-user-password")
	if password != "" {
		taskConfig.LocalUserPassword = c.FormValue("local-user-password")
	}

	taskConfig.LocalUserPrimaryGroup = c.FormValue("local-user-group")
	taskConfig.LocalUserSupplementaryGroup = c.FormValue("local-user-groups")
	taskConfig.LocalUserHome = c.FormValue("local-user-home")
	taskConfig.LocalUserShell = c.FormValue("local-user-shell")
	taskConfig.LocalUserUmask = c.FormValue("local-user-umask")

	confirmPassword := c.FormValue("local-user-password-confirm")
	if confirmPassword != taskConfig.LocalUserPassword {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_password_match"))
	}

	generateSSHKey := c.FormValue("local-user-generate-ssh-key")
	if generateSSHKey == "on" {
		taskConfig.LocalUserGenerateSSHKey = true
	}

	createHome := c.FormValue("local-user-create-home")
	if createHome == "on" {
		taskConfig.LocalUserCreateHome = true
	}

	skel := c.FormValue("local-user-skeleton")
	if skel != "" && !taskConfig.LocalUserCreateHome {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_skel_requires_home"))
	}
	taskConfig.LocalUserSkeleton = skel

	systemAccount := c.FormValue("local-user-system")
	if systemAccount == "on" {
		taskConfig.LocalUserSystemAccount = true
	}

	lockPassword := c.FormValue("local-user-password-lock")
	if lockPassword == "on" {
		taskConfig.LocalUserPasswordLock = true
	}

	localUID := c.FormValue("local-user-id")
	if localUID != "" {
		uid, err := strconv.Atoi(localUID)

		if err != nil || uid < 0 {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_uid_integer"))
		}
		taskConfig.LocalUserID = localUID
	}

	expires := c.FormValue("local-user-expires")
	if expires != "" {
		expiresTime, err := time.ParseInLocation("2006-01-02", expires, time.Local)
		if err != nil {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_could_not_parse_expire"))
		}
		taskConfig.LocalUserExpires = fmt.Sprintf("%d", expiresTime.Unix())
	}

	expireMax := c.FormValue("local-user-password-expire-max")
	if expireMax != "" {
		num, err := strconv.Atoi(expireMax)
		if err != nil || num < 0 {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_password_expire_max_not_valid"))
		}
	}
	taskConfig.LocalUserPasswordExpireMax = expireMax

	expireMin := c.FormValue("local-user-password-expire-min")
	if expireMin != "" {
		num, err := strconv.Atoi(expireMin)
		if err != nil || num < 0 {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_password_expire_min_not_valid"))
		}
	}
	taskConfig.LocalUserPasswordExpireMin = expireMin

	expireDisable := c.FormValue("local-user-password-expire-account-disable")
	if expireDisable != "" {
		num, err := strconv.Atoi(expireDisable)
		if err != nil || num < 0 {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_password_expire_account_disable_not_valid"))
		}
	}
	taskConfig.LocalUserPasswordExpireAccountDisable = expireDisable

	expireWarning := c.FormValue("local-user-password-expire-warn")
	if expireWarning != "" {
		num, err := strconv.Atoi(expireWarning)
		if err != nil || num < 0 {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_password_expire_warn_not_valid"))
		}
	}
	taskConfig.LocalUserPasswordExpireWarn = expireWarning

	sshKeyBits := c.FormValue("local-user-ssh-key-bits")
	if sshKeyBits != "" {
		num, err := strconv.Atoi(sshKeyBits)
		if err != nil || num < 0 {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_ssh_key_bits_not_valid"))
		}
	}
	taskConfig.LocalUserSSHKeyBits = sshKeyBits
	taskConfig.LocalUserSSHKeyComment = c.FormValue("local-user-ssh-key-comment")
	taskConfig.LocalUserSSHKeyFile = c.FormValue("local-user-ssh-key-file")
	taskConfig.LocalUserSSHKeyPassphrase = c.FormValue("local-user-ssh-key-passphrase")
	taskConfig.LocalUserSSHKeyType = c.FormValue("local-user-ssh-key-type")

	var uidMaxValue = 0
	uidMax := c.FormValue("local-user-uid-max")
	if uidMax != "" {
		uidMaxValue, err = strconv.Atoi(uidMax)
		if err != nil || uidMaxValue < 0 {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_uid_max_not_valid"))
		}
	}
	taskConfig.LocalUserUIDMax = uidMax

	var uidMinValue = 0
	uidMin := c.FormValue("local-user-uid-min")
	if uidMin != "" {
		uidMinValue, err = strconv.Atoi(uidMin)
		if err != nil || uidMinValue < 0 {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_uid_min_not_valid"))
		}
	}
	taskConfig.LocalUserUIDMin = uidMin

	if uidMin != "" && uidMax != "" {
		if uidMinValue >= uidMaxValue {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_max_min_error"))
		}
	}

	append := c.FormValue("local-user-append")
	if append == "on" {
		taskConfig.LocalUserAppend = true
	}

	overwriteSSHKey := c.FormValue("local-user-force")
	if overwriteSSHKey == "on" {
		taskConfig.LocalUserForce = true
	}

	return &taskConfig, nil
}

func validateRemoveLinuxLocalUser(c echo.Context) (*models.TaskConfig, error) {
	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	taskConfig.LocalUserUsername = c.FormValue("local-user-username")
	if taskConfig.LocalUserUsername == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_user_username_is_required"))
	}

	removeDirectories := c.FormValue("local-user-force")
	if removeDirectories == "on" {
		taskConfig.LocalUserForce = true
	}

	return &taskConfig, nil
}
func validateWindowsLocalGroup(c echo.Context) (*models.TaskConfig, error) {
	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

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

	return &taskConfig, nil
}

func validateUnixLocalGroup(c echo.Context) (*models.TaskConfig, error) {
	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	taskConfig.LocalGroupName = c.FormValue("local-unix-group-name")
	if (taskConfig.TaskType == "add_unix_local_group" || taskConfig.TaskType == "remove_unix_local_group") && taskConfig.LocalGroupName == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_group_name_is_required"))
	}

	taskConfig.LocalGroupDescription = c.FormValue("local-group-description")
	if taskConfig.TaskType == "add_local_group" && taskConfig.LocalGroupName == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_group_description_is_required"))
	}

	taskConfig.LocalGroupID = c.FormValue("local-group-id")
	if taskConfig.TaskType == "add_unix_local_group" && taskConfig.LocalGroupID != "" {
		if _, err := strconv.Atoi(taskConfig.LocalGroupID); err != nil {
			return nil, errors.New(i18n.T(c.Request().Context(), "tasks.local_gid_integer"))
		}
	}

	localGroupSystem := c.FormValue("local-group-system")
	if localGroupSystem == "on" {
		taskConfig.LocalGroupSystem = true
	}

	localGroupForce := c.FormValue("local-group-force")
	if localGroupForce == "on" {
		taskConfig.LocalGroupForce = true
	}

	return &taskConfig, nil
}

func validateMSI(c echo.Context) (*models.TaskConfig, error) {
	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

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

func validatePowerShellScript(c echo.Context) (*models.TaskConfig, error) {
	taskType := "powershell_script"
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	taskType = c.FormValue("task-type")
	taskConfig.ShellScript = c.FormValue("powershell-script")
	if taskType == "powershell_type" && taskConfig.ShellScript == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.shell_not_empty"))
	}

	taskConfig.ShellRunConfig = c.FormValue("powershell-run")
	if taskType == "powershell_type" && taskConfig.ShellRunConfig != task.ScriptRunAlways.String() && taskConfig.ShellRunConfig != task.ScriptRunOnce.String() {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.shell_wrong_run_config"))
	}

	return &taskConfig, nil
}

func validateUnixScript(c echo.Context) (*models.TaskConfig, error) {
	taskType := "unix_script"
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	taskType = c.FormValue("task-type")
	taskConfig.ShellScript = c.FormValue("unix-script")
	if taskType == "unix_script_type" && taskConfig.ShellScript == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.shell_not_empty"))
	}

	taskConfig.ShellExecute = c.FormValue("unix-script-executable")
	taskConfig.ShellCreates = c.FormValue("unix-script-creates")

	return &taskConfig, nil
}

func validateWinGetPackage(c echo.Context) (*models.TaskConfig, error) {
	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	taskConfig.PackageID = c.FormValue("package-id")
	if (taskConfig.TaskType == "winget_install" || taskConfig.TaskType == "winget_delete") && taskConfig.PackageID == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.package_id_not_empty"))
	}

	taskConfig.PackageName = c.FormValue("package-name")
	if (taskConfig.TaskType == "winget_install" || taskConfig.TaskType == "winget_delete") && taskConfig.PackageName == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.package_name_not_empty"))
	}

	return &taskConfig, nil
}

func validateFlatpakWinGetPackage(c echo.Context) (*models.TaskConfig, error) {
	taskType := c.FormValue("task-subtype")
	if c.FormValue("selected-task-type") != "" {
		taskType = c.FormValue("selected-task-type")
	}

	taskConfig := models.TaskConfig{
		TaskType:   taskType,
		AgentsType: c.FormValue("task-agent-type"),
	}

	taskConfig.Description = c.FormValue("task-description")
	if taskConfig.Description == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.new.empty"))
	}

	taskConfig.PackageID = c.FormValue("flatpak-id")
	if (taskConfig.TaskType == "flatpak_install" || taskConfig.TaskType == "flatpak_uninstall") && taskConfig.PackageID == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.package_id_not_empty"))
	}

	taskConfig.PackageName = c.FormValue("flatpak-name")
	if (taskConfig.TaskType == "flatpak_install" || taskConfig.TaskType == "flatpak_uninstall") && taskConfig.PackageName == "" {
		return nil, errors.New(i18n.T(c.Request().Context(), "tasks.package_name_not_empty"))
	}

	latest := c.FormValue("flatpak-latest")
	if latest == "on" {
		taskConfig.PackageLatest = true
	}

	return &taskConfig, nil
}
