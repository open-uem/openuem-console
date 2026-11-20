package models

import (
	"context"
	"strconv"
	"sync"
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

// saveInstallMu serializes saving install requests, see SaveDeployInfo
var saveInstallMu sync.Mutex

// mergeDuplicateDeployments reports whether an agent has a deployment of a package, and
// removes all but the oldest if there are several (consoles before this fix created one
// per install request)
func (m *Model) mergeDuplicateDeployments(packageID, agentID string) (bool, error) {
	ids, err := m.Client.Deployment.Query().
		Where(deployment.PackageID(packageID), deployment.HasOwnerWith(agent.ID(agentID))).
		Order(ent.Asc(deployment.FieldID)).
		IDs(context.Background())
	if err != nil {
		return false, err
	}
	if len(ids) > 1 {
		if _, err := m.Client.Deployment.Delete().Where(deployment.IDIn(ids[1:]...)).Exec(context.Background()); err != nil {
			return false, err
		}
	}
	return len(ids) > 0, nil
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
		// Two requests for the same package must not both see no deployment and create
		// one each
		saveInstallMu.Lock()
		defer saveInstallMu.Unlock()

		// A package installed again, e.g. after a failed or removed install, reuses its
		// deployment: the agent's result is saved to the single deployment of the
		// package and agent, so a second one would never get the result
		existing, err := m.mergeDuplicateDeployments(data.PackageId, data.AgentId)
		if err != nil {
			return err
		}

		if deploymentFailed || existing {
			query := m.Client.Deployment.Update().SetInstalled(timeZero).SetUpdated(timeZero).SetFailed(false)

			if data.PackageBranch != "" {
				query.SetBranch(data.PackageBranch)
			}

			if data.PackageBrewType != "" {
				query.SetBrewType(data.PackageBrewType)
			}

			if data.PackageVerified {
				query.SetVerified(true)
			}

			if siteID == -1 {
				return query.
					Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).
					Exec(context.Background())
			} else {
				return query.
					Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).
					Exec(context.Background())
			}
		} else {
			query := m.Client.Deployment.Create().
				SetInstalled(timeZero).
				SetFailed(false).
				SetUpdated(timeZero).
				SetPackageID(data.PackageId).
				SetName(data.PackageName).
				SetVersion(data.PackageVersion).
				SetOwnerID(data.AgentId)

			if data.PackageBranch != "" {
				query.SetBranch(data.PackageBranch)
			}

			if data.PackageBrewType != "" {
				query.SetBrewType(data.PackageBrewType)
			}

			if data.PackageVerified {
				query.SetVerified(true)
			}

			return query.Exec(context.Background())
		}
	}

	if data.Action == "update" {
		query := m.Client.Deployment.Update().
			SetUpdated(timeZero).
			SetFailed(false)

		if data.PackageBranch != "" {
			query.SetBranch(data.PackageBranch)
		}

		if data.PackageBrewType != "" {
			query.SetBrewType(data.PackageBrewType)
		}

		if data.PackageVerified {
			query.SetVerified(true)
		}

		if siteID == -1 {
			return query.
				Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).
				Exec(context.Background())
		} else {
			return query.
				Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).
				Exec(context.Background())
		}
	}

	if data.Action == "uninstall" {
		query := m.Client.Deployment.Update().
			SetInstalled(timeZero).
			SetFailed(false)

		if data.PackageBranch != "" {
			query.SetBranch(data.PackageBranch)
		}

		if data.PackageBrewType != "" {
			query.SetBrewType(data.PackageBrewType)
		}

		if data.PackageVerified {
			query.SetVerified(true)
		}

		if siteID == -1 {
			return query.
				Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))).
				Exec(context.Background())
		} else {
			return query.
				Where(deployment.And(deployment.PackageID(data.PackageId), deployment.HasOwnerWith(agent.ID(data.AgentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))).
				Exec(context.Background())
		}
	}

	return nil
}

func (m *Model) RemoveDeployment(id int) error {
	return m.Client.Deployment.DeleteOneID(id).Exec(context.Background())
}
