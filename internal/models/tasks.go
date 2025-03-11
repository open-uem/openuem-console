package models

import (
	"context"
	"fmt"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/profile"
	"github.com/open-uem/ent/task"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type TaskConfig struct {
	TaskType             string
	ExecuteCommand       string
	PackageID            string
	PackageName          string
	Description          string
	RegistryKey          string
	RegistryKeyValue     string
	RegistryKeyValueType string
	RegistryKeyValueData string
	RegistryHex          bool
	RegistryForce        bool
}

func (m *Model) CountAllTasksForProfile(profileID int) (int, error) {
	return m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID))).Count(context.Background())
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
	}
	return fmt.Errorf(i18n.T(c.Request().Context(), "tasks.unexpected_task_type"))
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
	}
	return fmt.Errorf(i18n.T(c.Request().Context(), "tasks.unexpected_task_type"))
}

func (m *Model) GetTasksForProfileByPage(p partials.PaginationAndSort, profileID int) ([]*ent.Task, error) {
	query := m.Client.Task.Query().Where(task.HasProfileWith(profile.ID(profileID)))

	return query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).Order(task.ByID()).All(context.Background())
}

func (m *Model) GetTasksById(taskId int) (*ent.Task, error) {
	return m.Client.Task.Query().WithProfile().Where(task.ID(taskId)).First(context.Background())
}

func (m *Model) DeleteTask(taskId int) error {
	return m.Client.Task.DeleteOneID(taskId).Exec(context.Background())
}
