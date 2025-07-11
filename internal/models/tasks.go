package models

import (
	"context"
	"errors"
	"strconv"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/profile"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/task"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type TaskConfig struct {
	TaskType                              string
	ExecuteCommand                        string
	PackageID                             string
	PackageName                           string
	PackageLatest                         bool
	Description                           string
	RegistryKey                           string
	RegistryKeyValue                      string
	RegistryKeyValueType                  string
	RegistryKeyValueData                  string
	RegistryHex                           bool
	RegistryForce                         bool
	LocalUserUsername                     string
	LocalUserDescription                  string
	LocalUserFullName                     string
	LocalUserPassword                     string
	LocalUserDisabled                     bool
	LocalUserPasswordChangeNotAllowed     bool
	LocalUserPasswordChangeRequired       bool
	LocalUserNeverExpires                 bool
	LocalUserID                           string
	LocalUserPrimaryGroup                 string
	LocalUserSupplementaryGroup           string
	LocalUserCreateHome                   bool
	LocalUserGenerateSSHKey               bool
	LocalUserSystemAccount                bool
	LocalUserHome                         string
	LocalUserShell                        string
	LocalUserUmask                        string
	LocalUserSkeleton                     string
	LocalUserExpires                      string
	LocalUserPasswordLock                 bool
	LocalUserPasswordExpireMax            string
	LocalUserPasswordExpireMin            string
	LocalUserPasswordExpireAccountDisable string
	LocalUserPasswordExpireWarn           string
	LocalUserSSHKeyBits                   string
	LocalUserSSHKeyComment                string
	LocalUserSSHKeyFile                   string
	LocalUserSSHKeyPassphrase             string
	LocalUserSSHKeyType                   string
	LocalUserUIDMax                       string
	LocalUserUIDMin                       string
	LocalUserForce                        bool
	LocalUserAppend                       bool
	LocalGroupName                        string
	LocalGroupDescription                 string
	LocalGroupMembers                     string
	LocalGroupMembersToInclude            string
	LocalGroupMembersToExclude            string
	LocalGroupID                          string
	LocalGroupSystem                      bool
	LocalGroupForce                       bool
	MsiProductID                          string
	MsiPath                               string
	MsiArguments                          string
	MsiLogPath                            string
	MsiHashAlgorithm                      string
	MsiFileHash                           string
	ShellScript                           string
	ShellRunConfig                        string
	ShellExecute                          string
	ShellCreates                          string
	AgentsType                            string
	HomeBrewUpgradeAll                    bool
	HomeBrewUpdate                        bool
	HomeBrewInstallOptions                string
	HomeBrewUpgradeOptions                string
	HomeBrewGreedy                        bool
}

func (m *Model) CountAllTasksForProfile(profileID int, c *partials.CommonInfo) (int, error) {

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return -1, err
	}

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return -1, err
	}

	if siteID == -1 {
		return -1, err
	}

	return m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID), profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
}

func (m *Model) AddTaskToProfile(c echo.Context, profileID int, cfg TaskConfig) error {
	switch cfg.TaskType {
	case "winget_install", "winget_delete":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).Exec(context.Background())
	case "add_registry_key":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).SetRegistryKey(cfg.RegistryKey).Exec(context.Background())
	case "remove_registry_key":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).SetRegistryKey(cfg.RegistryKey).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "update_registry_key_default_value":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetRegistryKey(cfg.RegistryKey).SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "add_registry_key_value":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).
			SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).
			SetRegistryHex(cfg.RegistryHex).
			SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "remove_registry_key_value":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).Exec(context.Background())
	case "add_local_user":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserDescription(cfg.LocalUserDescription).
			SetLocalUserFullname(cfg.LocalUserFullName).
			SetLocalUserPassword(cfg.LocalUserPassword).
			SetLocalUserDisable(cfg.LocalUserDisabled).
			SetLocalUserPasswordChangeNotAllowed(cfg.LocalUserPasswordChangeNotAllowed).
			SetLocalUserPasswordChangeRequired(cfg.LocalUserPasswordChangeRequired).
			SetLocalUserPasswordNeverExpires(cfg.LocalUserNeverExpires).
			Exec(context.Background())
	case "add_unix_local_user":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserDescription(cfg.LocalUserDescription).
			SetLocalUserGroup(cfg.LocalUserPrimaryGroup).
			SetLocalUserGroups(cfg.LocalUserSupplementaryGroup).
			SetLocalUserHome(cfg.LocalUserHome).
			SetLocalUserShell(cfg.LocalUserShell).
			SetLocalUserCreateHome(cfg.LocalUserCreateHome).
			SetLocalUserSkeleton(cfg.LocalUserSkeleton).
			SetLocalUserUmask(cfg.LocalUserUmask).
			SetLocalUserGenerateSSHKey(cfg.LocalUserGenerateSSHKey).
			SetLocalUserSystem(cfg.LocalUserSystemAccount).
			SetLocalUserPassword(cfg.LocalUserPassword).
			SetLocalUserID(cfg.LocalUserID).
			SetLocalUserExpires(cfg.LocalUserExpires).
			SetLocalUserPasswordLock(cfg.LocalUserPasswordLock).
			SetLocalUserPasswordExpireMax(cfg.LocalUserPasswordExpireMax).
			SetLocalUserPasswordExpireMin(cfg.LocalUserPasswordExpireMin).
			SetLocalUserPasswordExpireAccountDisable(cfg.LocalUserPasswordExpireAccountDisable).
			SetLocalUserPasswordExpireWarn(cfg.LocalUserPasswordExpireWarn).
			SetLocalUserSSHKeyBits(cfg.LocalUserSSHKeyBits).
			SetLocalUserSSHKeyComment(cfg.LocalUserSSHKeyComment).
			SetLocalUserSSHKeyFile(cfg.LocalUserSSHKeyFile).
			SetLocalUserSSHKeyPassphrase(cfg.LocalUserSSHKeyPassphrase).
			SetLocalUserSSHKeyType(cfg.LocalUserSSHKeyType).
			SetLocalUserIDMax(cfg.LocalUserUIDMax).
			SetLocalUserIDMin(cfg.LocalUserUIDMin).
			SetLocalUserForce(cfg.LocalUserForce).
			SetLocalUserAppend(cfg.LocalUserAppend).
			Exec(context.Background())
	case "remove_unix_local_user":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserForce(cfg.LocalUserForce).
			Exec(context.Background())
	case "remove_local_user":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalUserUsername(cfg.LocalUserUsername).
			Exec(context.Background())
	case "add_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembers(cfg.LocalGroupMembers).
			Exec(context.Background())
	case "remove_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			Exec(context.Background())
	case "add_unix_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupID(cfg.LocalGroupID).
			SetLocalGroupSystem(cfg.LocalGroupSystem).
			Exec(context.Background())
	case "remove_unix_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupForce(cfg.LocalGroupForce).
			Exec(context.Background())
	case "add_users_to_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToInclude(cfg.LocalGroupMembersToInclude).
			Exec(context.Background())
	case "remove_users_from_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToExclude(cfg.LocalGroupMembersToExclude).
			Exec(context.Background())
	case "msi_install", "msi_uninstall":
		query := m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetMsiProductid(cfg.MsiProductID).
			SetMsiPath(cfg.MsiPath).
			SetMsiArguments(cfg.MsiArguments).
			SetMsiLogPath(cfg.MsiLogPath)

		if cfg.MsiHashAlgorithm != "" && cfg.MsiFileHash != "" {
			query = query.SetMsiFileHashAlg(task.MsiFileHashAlg(cfg.MsiHashAlgorithm)).SetMsiFileHash(cfg.MsiFileHash)
		}
		return query.Exec(context.Background())
	case "powershell_script":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetScript(cfg.ShellScript).SetScriptRun(task.ScriptRun(cfg.ShellRunConfig)).Exec(context.Background())
	case "unix_script":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetScript(cfg.ShellScript).SetScriptCreates(cfg.ShellCreates).SetScriptExecutable(cfg.ShellExecute).Exec(context.Background())
	case "flatpak_install", "flatpak_uninstall":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetPackageLatest(cfg.PackageLatest).Exec(context.Background())
	case task.TypeBrewCaskInstall.String(), task.TypeBrewCaskUninstall.String(), task.TypeBrewCaskUpgrade.String(),
		task.TypeBrewFormulaInstall.String(), task.TypeBrewFormulaUninstall.String(), task.TypeBrewFormulaUpgrade.String():
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).
			SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetBrewUpdate(cfg.HomeBrewUpdate).SetBrewGreedy(cfg.HomeBrewGreedy).
			SetBrewInstallOptions(cfg.HomeBrewInstallOptions).SetBrewUpgradeOptions(cfg.HomeBrewUpgradeOptions).SetBrewUpgradeAll(cfg.HomeBrewUpgradeAll).Exec(context.Background())
	}
	return errors.New(i18n.T(c.Request().Context(), "tasks.unexpected_task_type"))
}

func (m *Model) UpdateTaskToProfile(c echo.Context, taskID int, cfg TaskConfig) error {
	switch cfg.TaskType {
	case "winget_install", "winget_delete":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).Exec(context.Background())
	case "add_registry_key":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetRegistryKey(cfg.RegistryKey).Exec(context.Background())
	case "remove_registry_key":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetRegistryKey(cfg.RegistryKey).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "update_registry_key_default_value":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetRegistryKey(cfg.RegistryKey).SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "add_registry_key_value":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).
			SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).
			SetRegistryHex(cfg.RegistryHex).
			SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "remove_registry_key_value":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).Exec(context.Background())
	case "add_local_user":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserDescription(cfg.LocalUserDescription).
			SetLocalUserFullname(cfg.LocalUserFullName).
			SetLocalUserPassword(cfg.LocalUserPassword).
			SetLocalUserDisable(cfg.LocalUserDisabled).
			SetLocalUserPasswordChangeNotAllowed(cfg.LocalUserPasswordChangeNotAllowed).
			SetLocalUserPasswordChangeRequired(cfg.LocalUserPasswordChangeRequired).
			SetLocalUserPasswordNeverExpires(cfg.LocalUserNeverExpires).
			Exec(context.Background())
	case "add_unix_local_user":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserDescription(cfg.LocalUserDescription).
			SetLocalUserGroup(cfg.LocalUserPrimaryGroup).
			SetLocalUserGroups(cfg.LocalUserSupplementaryGroup).
			SetLocalUserHome(cfg.LocalUserHome).
			SetLocalUserShell(cfg.LocalUserShell).
			SetLocalUserCreateHome(cfg.LocalUserCreateHome).
			SetLocalUserSkeleton(cfg.LocalUserSkeleton).
			SetLocalUserUmask(cfg.LocalUserUmask).
			SetLocalUserGenerateSSHKey(cfg.LocalUserGenerateSSHKey).
			SetLocalUserSystem(cfg.LocalUserSystemAccount).
			SetLocalUserPassword(cfg.LocalUserPassword).
			SetLocalUserID(cfg.LocalUserID).
			SetLocalUserExpires(cfg.LocalUserExpires).
			SetLocalUserPasswordLock(cfg.LocalUserPasswordLock).
			SetLocalUserPasswordExpireMax(cfg.LocalUserPasswordExpireMax).
			SetLocalUserPasswordExpireMin(cfg.LocalUserPasswordExpireMin).
			SetLocalUserPasswordExpireAccountDisable(cfg.LocalUserPasswordExpireAccountDisable).
			SetLocalUserPasswordExpireWarn(cfg.LocalUserPasswordExpireWarn).
			SetLocalUserSSHKeyBits(cfg.LocalUserSSHKeyBits).
			SetLocalUserSSHKeyComment(cfg.LocalUserSSHKeyComment).
			SetLocalUserSSHKeyFile(cfg.LocalUserSSHKeyFile).
			SetLocalUserSSHKeyPassphrase(cfg.LocalUserSSHKeyPassphrase).
			SetLocalUserSSHKeyType(cfg.LocalUserSSHKeyType).
			SetLocalUserIDMax(cfg.LocalUserUIDMax).
			SetLocalUserIDMin(cfg.LocalUserUIDMin).
			SetLocalUserForce(cfg.LocalUserForce).
			SetLocalUserAppend(cfg.LocalUserAppend).
			Exec(context.Background())
	case "remove_unix_local_user":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserForce(cfg.LocalUserForce).
			Exec(context.Background())
	case "remove_local_user":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalUserUsername(cfg.LocalUserUsername).
			Exec(context.Background())
	case "add_local_group":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembers(cfg.LocalGroupMembers).
			Exec(context.Background())
	case "remove_local_group":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalGroupName(cfg.LocalGroupName).
			Exec(context.Background())
	case "add_unix_local_group":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupID(cfg.LocalGroupID).
			SetLocalGroupSystem(cfg.LocalGroupSystem).
			Exec(context.Background())
	case "remove_unix_local_group":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupForce(cfg.LocalGroupForce).
			Exec(context.Background())
	case "add_users_to_local_group":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToInclude(cfg.LocalGroupMembersToInclude).
			Exec(context.Background())
	case "remove_users_from_local_group":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToExclude(cfg.LocalGroupMembersToExclude).
			Exec(context.Background())
	case "msi_install", "msi_uninstall":
		query := m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).
			SetMsiProductid(cfg.MsiProductID).
			SetMsiPath(cfg.MsiPath).
			SetMsiArguments(cfg.MsiArguments).
			SetMsiLogPath(cfg.MsiLogPath)

		if cfg.MsiHashAlgorithm != "" && cfg.MsiFileHash != "" {
			query = query.SetMsiFileHashAlg(task.MsiFileHashAlg(cfg.MsiHashAlgorithm)).SetMsiFileHash(cfg.MsiFileHash)
		}

		return query.Exec(context.Background())
	case "powershell_script":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetScript(cfg.ShellScript).SetScriptRun(task.ScriptRun(cfg.ShellRunConfig)).Exec(context.Background())
	case "unix_script":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetScript(cfg.ShellScript).SetScriptCreates(cfg.ShellCreates).SetScriptExecutable(cfg.ShellExecute).Exec(context.Background())
	case "flatpak_install", "flatpak_uninstall":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetPackageLatest(cfg.PackageLatest).Exec(context.Background())
	case task.TypeBrewCaskInstall.String(), task.TypeBrewCaskUninstall.String(), task.TypeBrewCaskUpgrade.String(),
		task.TypeBrewFormulaInstall.String(), task.TypeBrewFormulaUninstall.String(), task.TypeBrewFormulaUpgrade.String():
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetPackageID(cfg.PackageID).
			SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetBrewUpdate(cfg.HomeBrewUpdate).SetBrewGreedy(cfg.HomeBrewGreedy).
			SetBrewInstallOptions(cfg.HomeBrewInstallOptions).SetBrewUpgradeOptions(cfg.HomeBrewUpgradeOptions).SetBrewUpgradeAll(cfg.HomeBrewUpgradeAll).Exec(context.Background())

	}
	return errors.New(i18n.T(c.Request().Context(), "tasks.unexpected_task_type"))
}

func (m *Model) GetTasksForProfileByPage(p partials.PaginationAndSort, profileID int, c *partials.CommonInfo) ([]*ent.Task, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		return nil, err
	}

	query := m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID), profile.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))

	return query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).Order(task.ByID()).All(context.Background())
}

func (m *Model) GetTasksById(taskId int) (*ent.Task, error) {
	return m.Client.Task.Query().WithProfile().Where(task.ID(taskId)).First(context.Background())
}

func (m *Model) DeleteTask(taskId int) error {
	return m.Client.Task.DeleteOneID(taskId).Exec(context.Background())
}
