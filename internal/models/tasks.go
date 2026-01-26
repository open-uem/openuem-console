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
	PackageVersion                        string
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
	NetbirdGroups                         string
	NetbirdAllowExtraDNSLabels            bool
	IgnoreErrors                          bool
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

	// common query
	query := m.Client.Task.Create().
		SetName(cfg.Description).
		SetType(task.Type(cfg.TaskType)).
		SetAgentType(task.AgentType(cfg.AgentsType)).
		SetProfileID(profileID).
		SetIgnoreErrors(cfg.IgnoreErrors)

	switch cfg.TaskType {
	case task.TypeWingetInstall.String(), task.TypeWingetDelete.String():
		return query.SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetPackageVersion(cfg.PackageVersion).SetPackageLatest(cfg.PackageLatest).Exec(context.Background())
	case task.TypeAddRegistryKey.String():
		return query.SetProfileID(profileID).SetRegistryKey(cfg.RegistryKey).Exec(context.Background())
	case task.TypeRemoveRegistryKey.String():
		return query.SetRegistryKey(cfg.RegistryKey).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case task.TypeUpdateRegistryKeyDefaultValue.String():
		return query.
			SetRegistryKey(cfg.RegistryKey).SetRegistryKeyValueType(task.RegistryKeyValueTypeString).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case task.TypeAddRegistryKeyValue.String():
		return query.
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).
			SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).
			SetRegistryHex(cfg.RegistryHex).
			SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case task.TypeRemoveRegistryKeyValue.String():
		return query.
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).Exec(context.Background())
	case task.TypeAddLocalUser.String():
		return query.
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserDescription(cfg.LocalUserDescription).
			SetLocalUserFullname(cfg.LocalUserFullName).
			SetLocalUserPassword(cfg.LocalUserPassword).
			SetLocalUserDisable(cfg.LocalUserDisabled).
			SetLocalUserPasswordChangeNotAllowed(cfg.LocalUserPasswordChangeNotAllowed).
			SetLocalUserPasswordChangeRequired(cfg.LocalUserPasswordChangeRequired).
			SetLocalUserPasswordNeverExpires(cfg.LocalUserNeverExpires).
			Exec(context.Background())
	case task.TypeAddUnixLocalUser.String():
		return query.
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
	case task.TypeRemoveUnixLocalUser.String():
		return query.
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserForce(cfg.LocalUserForce).
			Exec(context.Background())
	case task.TypeRemoveLocalUser.String():
		return query.
			SetLocalUserUsername(cfg.LocalUserUsername).
			Exec(context.Background())
	case task.TypeAddLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembers(cfg.LocalGroupMembers).
			Exec(context.Background())
	case task.TypeRemoveLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			Exec(context.Background())
	case task.TypeAddUnixLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupID(cfg.LocalGroupID).
			SetLocalGroupSystem(cfg.LocalGroupSystem).
			Exec(context.Background())
	case task.TypeRemoveUnixLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupForce(cfg.LocalGroupForce).
			Exec(context.Background())
	case task.TypeAddUsersToLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToInclude(cfg.LocalGroupMembersToInclude).
			Exec(context.Background())
	case task.TypeRemoveUsersFromLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToExclude(cfg.LocalGroupMembersToExclude).
			Exec(context.Background())
	case task.TypeMsiInstall.String(), task.TypeMsiUninstall.String():
		query := query.
			SetMsiProductid(cfg.MsiProductID).
			SetMsiPath(cfg.MsiPath).
			SetMsiArguments(cfg.MsiArguments).
			SetMsiLogPath(cfg.MsiLogPath)

		if cfg.MsiHashAlgorithm != "" && cfg.MsiFileHash != "" {
			query = query.SetMsiFileHashAlg(task.MsiFileHashAlg(cfg.MsiHashAlgorithm)).SetMsiFileHash(cfg.MsiFileHash)
		}
		return query.Exec(context.Background())
	case task.TypePowershellScript.String():
		return query.
			SetScript(cfg.ShellScript).SetScriptRun(task.ScriptRun(cfg.ShellRunConfig)).Exec(context.Background())
	case task.TypeUnixScript.String():
		return query.
			SetScript(cfg.ShellScript).SetScriptCreates(cfg.ShellCreates).SetScriptExecutable(cfg.ShellExecute).Exec(context.Background())
	case task.TypeFlatpakInstall.String(), task.TypeFlatpakUninstall.String():
		return query.SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetPackageLatest(cfg.PackageLatest).Exec(context.Background())
	case task.TypeBrewCaskInstall.String(), task.TypeBrewCaskUninstall.String(), task.TypeBrewCaskUpgrade.String(),
		task.TypeBrewFormulaInstall.String(), task.TypeBrewFormulaUninstall.String(), task.TypeBrewFormulaUpgrade.String():
		return query.
			SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetBrewUpdate(cfg.HomeBrewUpdate).SetBrewGreedy(cfg.HomeBrewGreedy).
			SetBrewInstallOptions(cfg.HomeBrewInstallOptions).SetBrewUpgradeOptions(cfg.HomeBrewUpgradeOptions).SetBrewUpgradeAll(cfg.HomeBrewUpgradeAll).Exec(context.Background())
	case task.TypeNetbirdInstall.String(), task.TypeNetbirdUninstall.String():
		return query.Exec(context.Background())
	case task.TypeNetbirdRegister.String():
		tenantID := c.Param("tenant")
		if tenantID == "" {
			return errors.New("tenant ID cannot be empty")
		}
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return errors.New("could not parse tenant ID as an int")
		}

		return m.Client.Task.Create().SetName(cfg.Description).SetTenant(id).SetNetbirdGroups(cfg.NetbirdGroups).SetNetbirdAllowExtraDNSLabels(cfg.NetbirdAllowExtraDNSLabels).SetType(task.Type(cfg.TaskType)).SetAgentType(task.AgentType(cfg.AgentsType)).SetProfileID(profileID).Exec(context.Background())
	}
	return errors.New(i18n.T(c.Request().Context(), "tasks.unexpected_task_type"))
}

func (m *Model) UpdateProfileTask(c echo.Context, taskID int, cfg TaskConfig) error {

	// common query
	query := m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetIgnoreErrors(cfg.IgnoreErrors)

	// Update version
	query.AddVersion(1)

	// Specify values to be updated

	switch cfg.TaskType {
	case task.TypeWingetInstall.String(), task.TypeWingetDelete.String():
		return query.SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetPackageVersion(cfg.PackageVersion).SetPackageLatest(cfg.PackageLatest).Exec(context.Background())
	case task.TypeAddRegistryKey.String():
		return query.SetRegistryKey(cfg.RegistryKey).Exec(context.Background())
	case task.TypeRemoveRegistryKey.String():
		return query.SetRegistryKey(cfg.RegistryKey).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case task.TypeUpdateRegistryKeyDefaultValue.String():
		return query.SetRegistryKey(cfg.RegistryKey).SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case task.TypeAddRegistryKeyValue.String():
		return query.
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).
			SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).
			SetRegistryHex(cfg.RegistryHex).
			SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case task.TypeRemoveRegistryKeyValue.String():
		return query.
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).Exec(context.Background())
	case task.TypeAddLocalUser.String():
		return query.
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserDescription(cfg.LocalUserDescription).
			SetLocalUserFullname(cfg.LocalUserFullName).
			SetLocalUserPassword(cfg.LocalUserPassword).
			SetLocalUserDisable(cfg.LocalUserDisabled).
			SetLocalUserPasswordChangeNotAllowed(cfg.LocalUserPasswordChangeNotAllowed).
			SetLocalUserPasswordChangeRequired(cfg.LocalUserPasswordChangeRequired).
			SetLocalUserPasswordNeverExpires(cfg.LocalUserNeverExpires).
			Exec(context.Background())
	case task.TypeAddUnixLocalUser.String():
		return query.
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
	case task.TypeRemoveUnixLocalUser.String():
		return query.
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserForce(cfg.LocalUserForce).
			Exec(context.Background())
	case task.TypeRemoveLocalUser.String():
		return query.
			SetLocalUserUsername(cfg.LocalUserUsername).
			Exec(context.Background())
	case task.TypeAddLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembers(cfg.LocalGroupMembers).
			Exec(context.Background())
	case task.TypeRemoveLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			Exec(context.Background())
	case task.TypeAddUnixLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupID(cfg.LocalGroupID).
			SetLocalGroupSystem(cfg.LocalGroupSystem).
			Exec(context.Background())
	case task.TypeRemoveUnixLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupForce(cfg.LocalGroupForce).
			Exec(context.Background())
	case task.TypeAddUsersToLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToInclude(cfg.LocalGroupMembersToInclude).
			Exec(context.Background())
	case task.TypeRemoveUsersFromLocalGroup.String():
		return query.
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToExclude(cfg.LocalGroupMembersToExclude).
			Exec(context.Background())
	case task.TypeMsiInstall.String(), task.TypeMsiUninstall.String():
		query := query.
			SetMsiProductid(cfg.MsiProductID).
			SetMsiPath(cfg.MsiPath).
			SetMsiArguments(cfg.MsiArguments).
			SetMsiLogPath(cfg.MsiLogPath)

		if cfg.MsiHashAlgorithm != "" && cfg.MsiFileHash != "" {
			query = query.SetMsiFileHashAlg(task.MsiFileHashAlg(cfg.MsiHashAlgorithm)).SetMsiFileHash(cfg.MsiFileHash)
		}
		return query.Exec(context.Background())
	case task.TypePowershellScript.String():
		return query.SetScript(cfg.ShellScript).SetScriptRun(task.ScriptRun(cfg.ShellRunConfig)).Exec(context.Background())
	case task.TypeUnixScript.String():
		return query.SetScript(cfg.ShellScript).SetScriptCreates(cfg.ShellCreates).SetScriptExecutable(cfg.ShellExecute).Exec(context.Background())
	case task.TypeFlatpakInstall.String(), task.TypeFlatpakUninstall.String():
		return query.SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetPackageLatest(cfg.PackageLatest).Exec(context.Background())
	case task.TypeBrewCaskInstall.String(), task.TypeBrewCaskUninstall.String(), task.TypeBrewCaskUpgrade.String(),
		task.TypeBrewFormulaInstall.String(), task.TypeBrewFormulaUninstall.String(), task.TypeBrewFormulaUpgrade.String():
		return query.SetPackageID(cfg.PackageID).
			SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetBrewUpdate(cfg.HomeBrewUpdate).SetBrewGreedy(cfg.HomeBrewGreedy).
			SetBrewInstallOptions(cfg.HomeBrewInstallOptions).SetBrewUpgradeOptions(cfg.HomeBrewUpgradeOptions).SetBrewUpgradeAll(cfg.HomeBrewUpgradeAll).Exec(context.Background())
	case task.TypeNetbirdInstall.String(), task.TypeNetbirdUninstall.String():
		return query.Exec(context.Background())
	case task.TypeNetbirdRegister.String():
		tenantID := c.Param("tenant")
		if tenantID == "" {
			return errors.New("tenant ID cannot be empty")
		}
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return errors.New("could not parse tenant ID as an int")
		}

		return query.SetTenant(id).SetNetbirdGroups(cfg.NetbirdGroups).SetNetbirdAllowExtraDNSLabels(cfg.NetbirdAllowExtraDNSLabels).Exec(context.Background())
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

func (m *Model) GetTasksById(taskID int) (*ent.Task, error) {
	return m.Client.Task.Query().WithProfile().Where(task.ID(taskID)).First(context.Background())
}

func (m *Model) DeleteTask(taskID int) error {
	return m.Client.Task.DeleteOneID(taskID).Exec(context.Background())
}

func (m *Model) EnableTask(taskID int, disabled bool) error {
	return m.Client.Task.UpdateOneID(taskID).SetDisabled(disabled).Exec(context.Background())
}
