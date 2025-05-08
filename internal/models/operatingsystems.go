package models

import (
	"context"
	"strconv"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/operatingsystem"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) CountAgentsByOSVersion(c *partials.CommonInfo) ([]Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	// Info from agents waiting for admission won't be shown
	if siteID == -1 {
		agents := []Agent{}
		if err := m.Client.OperatingSystem.Query().Where(operatingsystem.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).GroupBy(operatingsystem.FieldVersion).Aggregate(ent.Count()).Scan(context.Background(), &agents); err != nil {
			return nil, err
		}
		return agents, err
	} else {
		agents := []Agent{}
		if err := m.Client.OperatingSystem.Query().Where(operatingsystem.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).GroupBy(operatingsystem.FieldVersion).Aggregate(ent.Count()).Scan(context.Background(), &agents); err != nil {
			return nil, err
		}
		return agents, err
	}
}

func (m *Model) GetOSVersions(f filters.AgentFilter, c *partials.CommonInfo) ([]string, error) {
	var query *ent.OperatingSystemSelect

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.OperatingSystem.Query().Where(operatingsystem.HasOwnerWith(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Unique(true).Select(operatingsystem.FieldVersion)
	} else {
		query = m.Client.OperatingSystem.Query().Where(operatingsystem.HasOwnerWith(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Unique(true).Select(operatingsystem.FieldVersion)
	}

	if len(f.AgentOSVersions) > 0 {
		query.Where(operatingsystem.TypeIn(f.AgentOSVersions...))
	}

	return query.Strings(context.Background())
}

func (m *Model) CountAllOSUsernames(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.OperatingSystem.Query().Select(operatingsystem.FieldUsername).Unique(true).Where(operatingsystem.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	} else {
		return m.Client.OperatingSystem.Query().Select(operatingsystem.FieldUsername).Unique(true).Where(operatingsystem.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	}
}
