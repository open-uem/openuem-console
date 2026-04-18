package models

import (
	"context"
	"errors"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/profile"
	"github.com/open-uem/ent/task"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type TaskConfig struct {
	TaskType                              string
	ExecuteCommand                        string
	PackageID                             string
	PackageName                           string
	PackageLatest                         bool
	PackageVersion                        string
	PackageBranch                         string
	PackageBrewType                       string
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

func (m *Model) CountAllTasksForProfile(profileID int) (int, error) {
	return m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID))).Count(context.Background())
}

func (m *Model) AddTaskToProfile(c echo.Context, profileID int, cfg TaskConfig) error {

	order := 0

	// let's see which is the highest order for tasks in profile
	t, err := m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID))).Order(task.ByOrder(sql.OrderDesc())).First(context.Background())
	if err == nil {
		order = t.Order
	}

	// common query
	query := m.Client.Task.Create().
		SetName(cfg.Description).
		SetType(task.Type(cfg.TaskType)).
		SetAgentType(task.AgentType(cfg.AgentsType)).
		SetProfileID(profileID).
		SetIgnoreErrors(cfg.IgnoreErrors).
		SetOrder(order + 1)

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
		return query.SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetPackageLatest(cfg.PackageLatest).SetPackageBranch(cfg.PackageBranch).Exec(context.Background())
	case task.TypeBrewCaskInstall.String(), task.TypeBrewCaskUninstall.String(), task.TypeBrewCaskUpgrade.String(),
		task.TypeBrewFormulaInstall.String(), task.TypeBrewFormulaUninstall.String(), task.TypeBrewFormulaUpgrade.String():
		return query.
			SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).SetBrewUpdate(cfg.HomeBrewUpdate).SetBrewGreedy(cfg.HomeBrewGreedy).SetPackageBrewType(cfg.PackageBrewType).
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

func (m *Model) GetTasksForProfileByPage(p partials.PaginationAndSort, profileID int) ([]*ent.Task, error) {
	// Check if we've values in the order column
	countWithOrder, err := m.Client.Task.Query().Where(task.OrderGT(0), task.HasProfileWith(profile.ID(profileID))).Count(context.Background())
	if err != nil {
		return nil, err
	}

	// If we don't have the order column filled with values let's add them
	if countWithOrder == 0 {
		// let's get all tasks we have
		tasks, err := m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID))).Order(task.ByID()).All(context.Background())
		if err != nil {
			return nil, err
		}

		// We must fill the order column as we're using it to order the results
		for i, t := range tasks {
			if err := m.Client.Task.UpdateOneID(t.ID).SetOrder(i + 1).Exec(context.Background()); err != nil {
				return nil, err
			}
		}
	}

	// Now, we have the ordered values, and we can use the order colum
	query := m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID)))

	return query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).Order(task.ByOrder()).All(context.Background())
}

func (m *Model) GetTasksById(taskID int) (*ent.Task, error) {
	return m.Client.Task.Query().WithProfile().Where(task.ID(taskID)).First(context.Background())
}

func (m *Model) DeleteTask(profileID int, taskID int) error {
	// get the curren task
	currentTask, err := m.Client.Task.Get(context.Background(), taskID)
	if err != nil {
		return err
	}

	// we must delete the task
	if err := m.Client.Task.DeleteOneID(taskID).Exec(context.Background()); err != nil {
		return err
	}

	//...but we must then update the order column from that column onwards
	return m.Client.Task.Update().Where(task.OrderGT(currentTask.Order)).AddOrder(-1).Exec(context.Background())
}

func (m *Model) EnableTask(taskID int, disabled bool) error {
	return m.Client.Task.UpdateOneID(taskID).SetDisabled(disabled).Exec(context.Background())
}

func (m *Model) MoveTask(taskID int, currentOrder int, newOrder int) error {
	t, err := m.Client.Task.Query().WithProfile().Where(task.ID(taskID)).Only(context.Background())
	if err != nil {
		return err
	}

	if currentOrder < newOrder {
		if err := m.Client.Task.Update().Where(
			task.HasProfileWith(profile.ID(t.Edges.Profile.ID)),
			task.OrderGTE(currentOrder),
			task.OrderLTE(newOrder),
		).AddOrder(-1).Exec(context.Background()); err != nil {
			return err
		}
	}

	if currentOrder > newOrder {
		if err := m.Client.Task.Update().Where(
			task.HasProfileWith(profile.ID(t.Edges.Profile.ID)),
			task.OrderGTE(newOrder),
			task.OrderLTE(currentOrder),
		).AddOrder(+1).Exec(context.Background()); err != nil {
			return err
		}
	}

	return m.Client.Task.Update().Where(task.ID(taskID)).SetOrder(newOrder).Exec(context.Background())
}

func (m *Model) CloneTask(taskID int, taskName string, profileID int, order int) error {
	t, err := m.Client.Task.Get(context.Background(), taskID)
	if err != nil {
		return err
	}

	query := m.Client.Task.Create()

	return CloneTask(query, t, taskName, profileID, order)
}

func (m *Model) CloneTaskInProfileTransaction(tx *ent.Tx, taskID int, taskName string, profileID int, order int) error {
	t, err := tx.Task.Get(context.Background(), taskID)
	if err != nil {
		return err
	}

	query := tx.Task.Create()

	return CloneTask(query, t, taskName, profileID, order)
}

func CloneTask(query *ent.TaskCreate, t *ent.Task, taskName string, profileID int, order int) error {
	query.SetAgentType(t.AgentType)
	query.SetAptAllowDowngrade(t.AptAllowDowngrade)
	query.SetAptDeb(t.AptDeb)
	query.SetAptDpkgOptions(t.AptDpkgOptions)
	query.SetAptFailOnAutoremove(t.AptFailOnAutoremove)
	query.SetAptForce(t.AptForce)
	query.SetAptInstallRecommends(t.AptInstallRecommends)
	query.SetAptName(taskName)
	query.SetAptOnlyUpgrade(t.AptOnlyUpgrade)
	query.SetAptPurge(t.AptPurge)
	query.SetAptUpdateCache(t.AptUpdateCache)
	query.SetAptUpgradeType(t.AptUpgradeType)
	query.SetBrewGreedy(t.BrewGreedy)
	query.SetBrewInstallOptions(t.BrewInstallOptions)
	query.SetBrewUpdate(t.BrewUpdate)
	query.SetBrewUpgradeAll(t.BrewUpgradeAll)
	query.SetBrewUpgradeOptions(t.BrewUpgradeOptions)
	query.SetDisabled(t.Disabled)
	query.SetIgnoreErrors(t.IgnoreErrors)
	query.SetLocalGroupDescription(t.LocalGroupDescription)
	query.SetLocalGroupForce(t.LocalGroupForce)
	query.SetLocalGroupID(t.LocalGroupID)
	query.SetLocalGroupMembers(t.LocalGroupMembers)
	query.SetLocalGroupMembersToExclude(t.LocalGroupMembersToExclude)
	query.SetLocalGroupMembersToInclude(t.LocalGroupMembersToInclude)
	query.SetLocalGroupName(t.LocalGroupName)
	query.SetLocalGroupSystem(t.LocalGroupSystem)
	query.SetLocalUserAppend(t.LocalUserAppend)
	query.SetLocalUserCreateHome(t.LocalUserCreateHome)
	query.SetLocalUserDescription(t.LocalUserDescription)
	query.SetLocalUserDisable(t.LocalUserDisable)
	query.SetLocalUserExpires(t.LocalUserExpires)
	query.SetLocalUserForce(t.LocalUserForce)
	query.SetLocalUserFullname(t.LocalUserFullname)
	query.SetLocalUserGenerateSSHKey(t.LocalUserGenerateSSHKey)
	query.SetLocalUserGroup(t.LocalUserGroup)
	query.SetLocalUserGroups(t.LocalUserGroups)
	query.SetLocalUserHome(t.LocalUserHome)
	query.SetLocalUserID(t.LocalUserID)
	query.SetLocalUserIDMax(t.LocalUserIDMax)
	query.SetLocalUserIDMin(t.LocalUserIDMin)
	query.SetLocalUserMoveHome(t.LocalUserMoveHome)
	query.SetLocalUserNonunique(t.LocalUserNonunique)
	query.SetLocalUserPassword(t.LocalUserPassword)
	query.SetLocalUserPasswordChangeNotAllowed(t.LocalUserPasswordChangeNotAllowed)
	query.SetLocalUserPasswordChangeRequired(t.LocalUserPasswordChangeRequired)
	query.SetLocalUserPasswordExpireAccountDisable(t.LocalUserPasswordExpireAccountDisable)
	query.SetLocalUserPasswordExpireMax(t.LocalUserPasswordExpireMax)
	query.SetLocalUserPasswordExpireMin(t.LocalUserPasswordExpireMin)
	query.SetLocalUserPasswordExpireWarn(t.LocalUserPasswordExpireWarn)
	query.SetLocalUserPasswordLock(t.LocalUserPasswordLock)
	query.SetLocalUserPasswordNeverExpires(t.LocalUserPasswordNeverExpires)
	query.SetLocalUserSSHKeyBits(t.LocalUserSSHKeyBits)
	query.SetLocalUserSSHKeyComment(t.LocalUserSSHKeyComment)
	query.SetLocalUserSSHKeyFile(t.LocalUserSSHKeyFile)
	query.SetLocalUserSSHKeyPassphrase(t.LocalUserSSHKeyPassphrase)
	query.SetLocalUserSSHKeyType(t.LocalUserSSHKeyType)
	query.SetLocalUserSeuser(t.LocalUserSeuser)
	query.SetLocalUserShell(t.LocalUserShell)
	query.SetLocalUserSkeleton(t.LocalUserSkeleton)
	query.SetLocalUserSystem(t.LocalUserSystem)
	query.SetLocalUserUmask(t.LocalUserUmask)
	query.SetLocalUserUsername(t.LocalUserUsername)
	query.SetMsiArguments(t.MsiArguments)
	query.SetMsiFileHash(t.MsiFileHash)

	if string(t.MsiFileHashAlg) != "" {
		query.SetMsiFileHashAlg(t.MsiFileHashAlg)
	}

	query.SetMsiLogPath(t.MsiLogPath)
	query.SetMsiPath(t.MsiPath)
	query.SetMsiProductid(t.MsiProductid)
	query.SetName(taskName).SetOrder(t.Order)
	query.SetPackageArch(t.PackageArch)
	query.SetPackageBranch(t.PackageBranch)
	query.SetPackageBrewType(t.PackageBrewType)
	query.SetPackageID(t.PackageID)
	query.SetPackageLatest(t.PackageLatest)
	query.SetPackageName(t.PackageName)
	query.SetPackageVersion(t.PackageVersion)
	query.SetProfileID(profileID)
	query.SetRegistryForce(t.RegistryForce)
	query.SetRegistryHex(t.RegistryHex)
	query.SetRegistryKey(t.RegistryKey)
	query.SetRegistryKeyValueData(t.RegistryKeyValueData)
	query.SetRegistryKeyValueName(t.RegistryKeyValueName)

	if string(t.RegistryKeyValueType) != "" {
		query.SetRegistryKeyValueType(t.RegistryKeyValueType)
	}

	query.SetScript(t.Script)
	query.SetScriptCreates(t.ScriptCreates)
	query.SetScriptExecutable(t.ScriptExecutable)

	if string(t.ScriptRun) != "" {
		query.SetScriptRun(t.ScriptRun)
	}

	query.SetType(t.Type)
	query.SetVersion(1)

	query.SetOrder(order)

	return query.Exec(context.Background())
}

func (m *Model) GetLasTaskOrderInProfile(profileID int) (*ent.Task, error) {
	return m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID))).Order(ent.Desc(task.FieldOrder)).First(context.Background())
}

func (m *Model) GetTaskSensitiveInformation() ([]*ent.Task, error) {
	return m.Client.Task.Query().Select(task.FieldID, task.FieldLocalUserPassword).All(context.Background())
}

func (m *Model) UpdateLocalUserPassword(taskID int, password string) error {
	return m.Client.Task.UpdateOneID(taskID).SetLocalUserPassword(password).Exec(context.Background())
}
