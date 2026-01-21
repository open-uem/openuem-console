package models

import (
	"context"
	"strconv"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/app"
	"github.com/open-uem/ent/computer"
	"github.com/open-uem/ent/operatingsystem"
	"github.com/open-uem/ent/predicate"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tag"
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
	var query *ent.OperatingSystemQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.OperatingSystem.Query().Where(operatingsystem.HasOwnerWith(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Unique(true)
	} else {
		query = m.Client.OperatingSystem.Query().Where(operatingsystem.HasOwnerWith(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Unique(true)
	}

	if len(f.AgentOSVersions) > 0 {
		query.Where(operatingsystem.TypeIn(f.AgentOSVersions...))
	}

	if len(f.Nickname) > 0 {
		query.Where(operatingsystem.HasOwnerWith(agent.NicknameContainsFold(f.Nickname)))
	}

	if len(f.Username) > 0 {
		query.Where(operatingsystem.UsernameContainsFold(f.Username))
	}

	if len(f.OSVersions) > 0 {
		query.Where(operatingsystem.VersionIn(f.OSVersions...))
	}

	if len(f.ComputerManufacturers) > 0 {
		query.Where(operatingsystem.HasOwnerWith(agent.HasComputerWith(computer.ManufacturerIn(f.ComputerManufacturers...))))
	}

	if len(f.ComputerModels) > 0 {
		query.Where(operatingsystem.HasOwnerWith(agent.HasComputerWith(computer.ModelIn(f.ComputerModels...))))
	}

	if len(f.WithApplication) > 0 && len(f.WithApplicationPublisher) > 0 {
		query.Where(operatingsystem.HasOwnerWith(agent.HasAppsWith(app.And(app.Name(f.WithApplication), app.Publisher(f.WithApplicationPublisher)))))
	} else {
		if len(f.WithApplication) > 0 {
			query.Where(operatingsystem.HasOwnerWith(agent.HasAppsWith(app.Name(f.WithApplication))))
		}
		if len(f.WithApplicationPublisher) > 0 {
			query.Where(operatingsystem.HasOwnerWith(agent.HasAppsWith(app.Name(f.WithApplicationPublisher))))
		}
	}

	if len(f.IsRemote) > 0 {
		if len(f.IsRemote) == 1 && f.IsRemote[0] == "Remote" {
			query.Where(operatingsystem.HasOwnerWith(agent.IsRemote(true)))
		}

		if len(f.IsRemote) == 1 && f.IsRemote[0] == "Local" {
			query.Where(operatingsystem.HasOwnerWith(agent.IsRemote(false)))
		}
	}

	if len(f.Tags) > 0 {
		predicates := []predicate.Agent{}
		for _, id := range f.Tags {
			predicates = append(predicates, agent.HasTagsWith(tag.ID(id)))
		}
		if len(predicates) > 0 {
			query.Where(operatingsystem.HasOwnerWith(agent.And(predicates...)))
		}
	}

	if len(f.Search) > 0 {
		query.Where(operatingsystem.HasOwnerWith(agent.Or(
			agent.NicknameContainsFold(f.Search),
			agent.OsIn(f.Search),
			agent.HasOperatingsystemWith(operatingsystem.UsernameContainsFold(f.Search)),
			agent.HasComputerWith(computer.ManufacturerContainsFold(f.Search)),
			agent.HasComputerWith(computer.ModelContainsFold(f.Search)),
		)))
	}

	return query.Select(operatingsystem.FieldVersion).Strings(context.Background())
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
