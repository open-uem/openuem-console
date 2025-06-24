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
	TaskType                          string
	ExecuteCommand                    string
	PackageID                         string
	PackageName                       string
	Description                       string
	RegistryKey                       string
	RegistryKeyValue                  string
	RegistryKeyValueType              string
	RegistryKeyValueData              string
	RegistryHex                       bool
	RegistryForce                     bool
	LocalUserUsername                 string
	LocalUserDescription              string
	LocalUserFullName                 string
	LocalUserPassword                 string
	LocalUserDisabled                 bool
	LocalUserPasswordChangeNotAllowed bool
	LocalUserPasswordChangeRequired   bool
	LocalUserNeverExpires             bool
	LocalGroupName                    string
	LocalGroupDescription             string
	LocalGroupMembers                 string
	LocalGroupMembersToInclude        string
	LocalGroupMembersToExclude        string
	MsiProductID                      string
	MsiPath                           string
	MsiArguments                      string
	MsiLogPath                        string
	MsiHashAlgorithm                  string
	MsiFileHash                       string
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
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).Exec(context.Background())
	case "add_registry_key":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).SetRegistryKey(cfg.RegistryKey).Exec(context.Background())
	case "remove_registry_key":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).SetRegistryKey(cfg.RegistryKey).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "update_registry_key_default_value":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetRegistryKey(cfg.RegistryKey).SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "add_registry_key_value":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).
			SetRegistryKeyValueType(task.RegistryKeyValueType(cfg.RegistryKeyValueType)).
			SetRegistryKeyValueData(cfg.RegistryKeyValueData).
			SetRegistryHex(cfg.RegistryHex).
			SetRegistryForce(cfg.RegistryForce).Exec(context.Background())
	case "remove_registry_key_value":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetRegistryKey(cfg.RegistryKey).
			SetRegistryKeyValueName(cfg.RegistryKeyValue).Exec(context.Background())
	case "add_local_user":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetLocalUserUsername(cfg.LocalUserUsername).
			SetLocalUserDescription(cfg.LocalUserDescription).
			SetLocalUserFullname(cfg.LocalUserFullName).
			SetLocalUserPassword(cfg.LocalUserPassword).
			SetLocalUserDisable(cfg.LocalUserDisabled).
			SetLocalUserPasswordChangeNotAllowed(cfg.LocalUserPasswordChangeNotAllowed).
			SetLocalUserPasswordChangeRequired(cfg.LocalUserPasswordChangeRequired).
			SetLocalUserPasswordNeverExpires(cfg.LocalUserNeverExpires).
			Exec(context.Background())
	case "remove_local_user":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetLocalUserUsername(cfg.LocalUserUsername).
			Exec(context.Background())
	case "add_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembers(cfg.LocalGroupMembers).
			Exec(context.Background())
	case "remove_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			Exec(context.Background())
	case "add_users_to_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToInclude(cfg.LocalGroupMembersToInclude).
			Exec(context.Background())
	case "remove_users_from_local_group":
		return m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetLocalGroupName(cfg.LocalGroupName).
			SetLocalGroupDescription(cfg.LocalGroupDescription).
			SetLocalGroupMembersToExclude(cfg.LocalGroupMembersToExclude).
			Exec(context.Background())
	case "msi_install", "msi_uninstall":
		query := m.Client.Task.Create().SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetProfileID(profileID).
			SetMsiProductid(cfg.MsiProductID).
			SetMsiPath(cfg.MsiPath).
			SetMsiArguments(cfg.MsiArguments).
			SetMsiLogPath(cfg.MsiLogPath)

		if cfg.MsiHashAlgorithm != "" && cfg.MsiFileHash != "" {
			query = query.SetMsiFileHashAlg(task.MsiFileHashAlg(cfg.MsiHashAlgorithm)).SetMsiFileHash(cfg.MsiFileHash)
		}
		return query.Exec(context.Background())
	}
	return errors.New(i18n.T(c.Request().Context(), "tasks.unexpected_task_type"))
}

func (m *Model) UpdateTaskToProfile(c echo.Context, taskID int, cfg TaskConfig) error {
	switch cfg.TaskType {
	case "winget_install", "winget_delete":
		return m.Client.Task.UpdateOneID(taskID).SetName(cfg.Description).SetType(task.Type(cfg.TaskType)).SetPackageID(cfg.PackageID).SetPackageName(cfg.PackageName).Exec(context.Background())
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
