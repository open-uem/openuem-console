package models

import (
	"context"
	"strconv"
	"time"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/deployment"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) GetDeploymentsForAgent(agentId string, p partials.PaginationAndSort, c *partials.CommonInfo) ([]*ent.Deployment, error) {
	var query *ent.DeploymentQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

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

func (m *Model) CountDeploymentsForAgent(agentId string, c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	} else {
		return m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	}
}

func (m *Model) GetDeployment(agentId, packageId string, c *partials.CommonInfo) (*ent.Deployment, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		return m.Client.Deployment.Query().Where(deployment.And(deployment.PackageID(packageId), deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).First(context.Background())
	} else {
		return m.Client.Deployment.Query().Where(deployment.And(deployment.PackageID(packageId), deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).First(context.Background())
	}
}

func (m *Model) DeploymentFailed(agentId, packageId string, c *partials.CommonInfo) (bool, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return false, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return false, err
	}

	if siteID == -1 {
		return m.Client.Deployment.Query().Where(deployment.And(deployment.PackageID(packageId), deployment.FailedEQ(true), deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).Exist(context.Background())
	} else {
		return m.Client.Deployment.Query().Where(deployment.And(deployment.PackageID(packageId), deployment.FailedEQ(true), deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).Exist(context.Background())
	}
}

func (m *Model) DeploymentAlreadyInstalled(agentId, packageId string, c *partials.CommonInfo) (bool, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return false, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return false, err
	}

	if siteID == -1 {
		return m.Client.Deployment.Query().Where(deployment.And(deployment.PackageID(packageId), deployment.InstalledNEQ(time.Time{}), deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).Exist(context.Background())
	} else {
		return m.Client.Deployment.Query().Where(deployment.And(deployment.PackageID(packageId), deployment.InstalledNEQ(time.Time{}), deployment.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).Exist(context.Background())
	}
}

func (m *Model) CountAllDeployments(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	} else {
		return m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	}
}

func (m *Model) SaveDeployInfo(data *openuem_nats.DeployAction, deploymentFailed bool, c *partials.CommonInfo) error {
	timeZero := time.Date(0001, 1, 1, 00, 00, 00, 00, time.UTC)

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if data.Action == "install" {
		if deploymentFailed {
			if siteID == -1 {
				return m.Client.Deployment.Update().
					SetInstalled(timeZero).
					SetUpdated(timeZero).
					SetFailed(false).
					Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).
					Exec(context.Background())
			} else {
				return m.Client.Deployment.Update().
					SetInstalled(timeZero).
					SetUpdated(timeZero).
					SetFailed(false).
					Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).
					Exec(context.Background())
			}
		} else {
			return m.Client.Deployment.Create().
				SetInstalled(timeZero).
				SetFailed(false).
				SetUpdated(timeZero).
				SetPackageID(data.PackageId).
				SetName(data.PackageName).
				SetVersion(data.PackageVersion).
				SetOwnerID(data.AgentId).
				Exec(context.Background())
		}
	}

	if data.Action == "update" {
		if siteID == -1 {
			return m.Client.Deployment.Update().
				SetUpdated(timeZero).
				SetFailed(false).
				Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).
				Exec(context.Background())
		} else {
			return m.Client.Deployment.Update().
				SetUpdated(timeZero).
				SetFailed(false).
				Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).
				Exec(context.Background())
		}
	}

	if data.Action == "uninstall" {
		if siteID == -1 {
			return m.Client.Deployment.Update().
				SetInstalled(timeZero).
				SetFailed(false).
				Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).
				Exec(context.Background())
		} else {
			return m.Client.Deployment.Update().
				SetInstalled(timeZero).
				SetFailed(false).
				Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).
				Exec(context.Background())
		}
	}

	return nil
}

func (m *Model) RemoveDeployment(id int) error {
	return m.Client.Deployment.DeleteOneID(id).Exec(context.Background())
}
