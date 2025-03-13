package models

import (
	"context"
	"time"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/deployment"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) GetDeploymentsForAgent(agentId string, p partials.PaginationAndSort) ([]*ent.Deployment, error) {
	query := m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.ID(agentId)))

	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(deployment.FieldName))
		} else {
			query = query.Order(ent.Desc(deployment.FieldName))
		}
	case "installation":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(deployment.FieldInstalled))
		} else {
			query = query.Order(ent.Desc(deployment.FieldInstalled))
		}
	case "updated":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(deployment.FieldUpdated))
		} else {
			query = query.Order(ent.Desc(deployment.FieldUpdated))
		}
	default:
		query = query.Order(ent.Desc(deployment.FieldInstalled))
	}

	deployments, err := query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

func (m *Model) CountDeploymentsForAgent(agentId string) (int, error) {
	return m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.ID(agentId))).Count(context.Background())
}

func (m *Model) DeploymentAlreadyInstalled(agentId, packageId string) (bool, error) {
	return m.Client.Deployment.Query().Where(deployment.And(deployment.PackageID(packageId), deployment.HasOwnerWith(agent.ID(agentId)))).Exist(context.Background())
}

func (m *Model) CountAllDeployments() (int, error) {
	return m.Client.Deployment.Query().Count(context.Background())
}

func (m *Model) SaveDeployInfo(data *openuem_nats.DeployAction) error {
	timeZero := time.Date(0001, 1, 1, 00, 00, 00, 00, time.UTC)

	if data.Action == "install" {
		return m.Client.Deployment.Create().
			SetInstalled(timeZero).
			SetUpdated(timeZero).
			SetPackageID(data.PackageId).
			SetName(data.PackageName).
			SetVersion(data.PackageVersion).
			SetOwnerID(data.AgentId).
			Exec(context.Background())
	}

	if data.Action == "update" {
		return m.Client.Deployment.Update().
			SetUpdated(timeZero).
			Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId)))).
			Exec(context.Background())
	}

	if data.Action == "uninstall" {
		return m.Client.Deployment.Update().
			SetInstalled(timeZero).
			Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId)))).
			Exec(context.Background())
	}

	return nil
}
