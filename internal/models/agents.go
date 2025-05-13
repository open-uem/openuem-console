package models

import (
	"context"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/antivirus"
	"github.com/open-uem/ent/predicate"
	"github.com/open-uem/ent/release"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/systemupdate"
	"github.com/open-uem/ent/tag"
	"github.com/open-uem/ent/tenant"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type Agent struct {
	OS      string
	Version string
	Status  string
	Count   int
}

func (m *Model) GetAllAgents(f filters.AgentFilter, c *partials.CommonInfo) ([]*ent.Agent, error) {
	var query *ent.AgentQuery

	// Info from agents waiting for admission won't be shown
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Agent.Query().WithRelease().Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = m.Client.Agent.Query().WithRelease().Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	// Apply filters
	applyAgentFilters(query, f)

	agents, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (m *Model) GetAgentsByPage(p partials.PaginationAndSort, f filters.AgentFilter, excludeWaitingForAdmissionAgents bool, c *partials.CommonInfo) ([]*ent.Agent, error) {
	var err error
	var agents []*ent.Agent
	var query *ent.AgentQuery

	// Info from agents waiting for admission won't be shown
	if excludeWaitingForAdmissionAgents {
		query = m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).WithSite().WithTags().WithRelease()
	} else {
		query = m.Client.Agent.Query().WithSite().WithTags().WithRelease()
	}

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = query.Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = query.Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	if p.PageSize != 0 {
		query = query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
	}

	// Apply filters
	applyAgentFilters(query, f)

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldHostname)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldHostname)).All(context.Background())
		}
	case "os":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldOs)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldOs)).All(context.Background())
		}
	case "version":
		if p.SortOrder == "asc" {
			agents, err = query.Order(agent.ByReleaseField(release.FieldVersion, sql.OrderAsc())).All(context.Background())
		} else {
			agents, err = query.Order(agent.ByReleaseField(release.FieldVersion, sql.OrderDesc())).All(context.Background())
		}
	case "last_contact":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldLastContact)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldLastContact)).All(context.Background())
		}
	case "status":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldAgentStatus)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldAgentStatus)).All(context.Background())
		}
	case "ip_address":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldIP)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldIP)).All(context.Background())
		}
	case "remote":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldIsRemote)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldIsRemote)).All(context.Background())
		}
	default:
		agents, err = query.Order(ent.Desc(agent.FieldLastContact)).All(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (m *Model) GetAgentById(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithTags().WithComputer().WithOperatingsystem().WithSite().Where(agent.ID(agentId)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, err
	} else {
		agent, err := m.Client.Agent.Query().WithTags().WithComputer().WithOperatingsystem().WithSite().Where(agent.ID(agentId)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, err
	}
}

func (m *Model) GetAgentOverviewById(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithSite().WithTags().WithComputer().WithOperatingsystem().WithAntivirus().WithSystemupdate().WithRelease().Where(agent.ID(agentId)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, err
	} else {
		agent, err := m.Client.Agent.Query().WithSite().WithTags().WithComputer().WithOperatingsystem().WithAntivirus().WithSystemupdate().WithRelease().Where(agent.ID(agentId)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, err
	}
}

func (m *Model) CountAgentsByOS(c *partials.CommonInfo) ([]Agent, error) {

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	// Info from agents waiting for admission won't be shown
	agents := []Agent{}

	if siteID == -1 {
		if err = m.Client.Agent.Query().Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Modify(func(s *sql.Selector) {
			s.Select(agent.FieldOs, sql.As(sql.Count("os"), "count")).Where(sql.And(sql.NEQ(agent.FieldAgentStatus, agent.AgentStatusWaitingForAdmission))).GroupBy("os").OrderBy("count")
		}).Scan(context.Background(), &agents); err != nil {
			return nil, err
		}
		return agents, err
	} else {
		if err = m.Client.Agent.Query().Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Modify(func(s *sql.Selector) {
			s.Select(agent.FieldOs, sql.As(sql.Count("os"), "count")).Where(sql.And(sql.NEQ(agent.FieldAgentStatus, agent.AgentStatusWaitingForAdmission))).GroupBy("os").OrderBy("count")
		}).Scan(context.Background(), &agents); err != nil {
			return nil, err
		}
		return agents, err
	}
}

func (m *Model) CountAllAgents(f filters.AgentFilter, excludeWaitingForAdmissionAgents bool, c *partials.CommonInfo) (int, error) {
	var query *ent.AgentQuery

	// Info from agents waiting for admission won't be shown
	if excludeWaitingForAdmissionAgents {
		query = m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))
	} else {
		query = m.Client.Agent.Query()
	}

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return -1, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return -1, err
	}

	if siteID == -1 {
		query = query.Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = query.Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	applyAgentFilters(query, f)

	count, err := query.Count(context.Background())
	return count, err
}

func (m *Model) GetAgentsUsedOSes(c *partials.CommonInfo) ([]string, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Unique(true).Select(agent.FieldOs).Strings(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Unique(true).Select(agent.FieldOs).Strings(context.Background())
	}
}

func applyAgentFilters(query *ent.AgentQuery, f filters.AgentFilter) {
	if len(f.Hostname) > 0 {
		query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if len(f.AgentStatusOptions) > 0 {
		if len(f.AgentStatusOptions) == 1 && f.AgentStatusOptions[0] == "WaitingForAdmission" {
			query.Where(agent.AgentStatusEQ(agent.AgentStatusWaitingForAdmission))
		}

		if len(f.AgentStatusOptions) == 1 && f.AgentStatusOptions[0] == "Enabled" {
			query.Where(agent.AgentStatusEQ(agent.AgentStatusEnabled))
		}

		if len(f.AgentStatusOptions) == 1 && f.AgentStatusOptions[0] == "No Contact" {
			query.Where(agent.AgentStatusEQ(agent.AgentStatusEnabled))
		}

		if len(f.AgentStatusOptions) == 1 && f.AgentStatusOptions[0] == "Disabled" {
			query.Where(agent.AgentStatusEQ(agent.AgentStatusDisabled))
		}
	}

	if len(f.IsRemote) > 0 {
		if len(f.IsRemote) == 1 && f.IsRemote[0] == "Remote" {
			query.Where(agent.IsRemote(true))
		}

		if len(f.IsRemote) == 1 && f.IsRemote[0] == "Local" {
			query.Where(agent.IsRemote(false))
		}
	}

	if len(f.AgentOSVersions) > 0 {
		query.Where(agent.OsIn(f.AgentOSVersions...))
	}

	if len(f.ContactFrom) > 0 {
		from, err := time.Parse("2006-01-02", f.ContactFrom)
		if err == nil {
			query.Where(agent.LastContactGTE(from))
		}
	}

	if len(f.ContactTo) > 0 {
		to, err := time.Parse("2006-01-02", f.ContactTo)
		if err == nil {
			query.Where(agent.LastContactLTE(to))
		}
	}

	if len(f.Tags) > 0 {
		predicates := []predicate.Agent{}
		for _, id := range f.Tags {
			predicates = append(predicates, agent.HasTagsWith(tag.ID(id)))
		}
		if len(predicates) > 0 {
			query.Where(agent.And(predicates...))
		}
	}

	if f.NoContact {
		query.Where(agent.LastContactLTE((time.Now().AddDate(0, 0, -1))))
	}
}

func (m *Model) CountAgentsReportedLast24h(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		count, err := m.Client.Agent.Query().Where(agent.LastContactGTE(time.Now().AddDate(0, 0, -1)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
		if err != nil {
			return 0, err
		}
		return count, err
	} else {
		count, err := m.Client.Agent.Query().Where(agent.LastContactGTE(time.Now().AddDate(0, 0, -1)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
		if err != nil {
			return 0, err
		}
		return count, err
	}
}

func (m *Model) CountAgentsNotReportedLast24h(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		count, err := m.Client.Agent.Query().Where(agent.LastContactLT(time.Now().AddDate(0, 0, -1)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
		if err != nil {
			return 0, err
		}
		return count, err
	} else {
		count, err := m.Client.Agent.Query().Where(agent.LastContactLT(time.Now().AddDate(0, 0, -1)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
		if err != nil {
			return 0, err
		}
		return count, err
	}
}

func (m *Model) DeleteAgent(agentId string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		err = m.Client.Agent.DeleteOneID(agentId).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
		if err != nil {
			return err
		}
	} else {
		err = m.Client.Agent.DeleteOneID(agentId).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) EnableAgent(agentId string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		if _, err := m.Client.Agent.UpdateOneID(agentId).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).SetAgentStatus(agent.AgentStatusEnabled).Save(context.Background()); err != nil {
			return err
		}
	} else {
		if _, err := m.Client.Agent.UpdateOneID(agentId).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).SetAgentStatus(agent.AgentStatusEnabled).Save(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func (m *Model) DisableAgent(agentId string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		if _, err := m.Client.Agent.UpdateOneID(agentId).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).SetAgentStatus(agent.AgentStatusDisabled).Save(context.Background()); err != nil {
			return err
		}
	} else {
		if _, err := m.Client.Agent.UpdateOneID(agentId).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).SetAgentStatus(agent.AgentStatusDisabled).Save(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) AddTagToAgent(agentId, tagId string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(tagId)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return m.Client.Agent.UpdateOneID(agentId).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).AddTagIDs(id).Exec(context.Background())
	} else {
		return m.Client.Agent.UpdateOneID(agentId).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).AddTagIDs(id).Exec(context.Background())
	}
}

func (m *Model) RemoveTagFromAgent(agentId, tagId string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(tagId)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return m.Client.Agent.UpdateOneID(agentId).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).RemoveTagIDs(id).Exec(context.Background())
	} else {
		return m.Client.Agent.UpdateOneID(agentId).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).RemoveTagIDs(id).Exec(context.Background())
	}
}

func (m *Model) CountPendingUpdateAgents(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.HasSystemupdateWith(systemupdate.PendingUpdatesEQ(true)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.HasSystemupdateWith(systemupdate.PendingUpdatesEQ(true)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	}
}

func (m *Model) CountDisabledAntivirusAgents(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.HasAntivirusWith(antivirus.IsActive(false)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.Os("windows")).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.HasAntivirusWith(antivirus.IsActive(false)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.Os("windows")).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	}
}

func (m *Model) CountOutdatedAntivirusDatabaseAgents(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.HasAntivirusWith(antivirus.IsUpdated(false)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.Os("windows")).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.HasAntivirusWith(antivirus.IsUpdated(false)), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.Os("windows")).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	}
}

func (m *Model) CountNoAutoupdateAgents(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.HasSystemupdateWith(systemupdate.Not(systemupdate.SystemUpdateStatusContains(openuem_nats.NOTIFY_SCHEDULED_INSTALLATION))), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.HasSystemupdateWith(systemupdate.Not(systemupdate.SystemUpdateStatusContains(openuem_nats.NOTIFY_SCHEDULED_INSTALLATION))), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	}
}

func (m *Model) CountVNCSupportedAgents(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.Not(agent.Vnc("")), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.Not(agent.Vnc("")), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	}
}

func (m *Model) CountDisabledAgents(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.AgentStatusEQ(agent.AgentStatusDisabled)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.AgentStatusEQ(agent.AgentStatusDisabled)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	}
}

func (m *Model) CountWaitingForAdmissionAgents(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.AgentStatusEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.AgentStatusEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Count(context.Background())
	}
}

func (m *Model) AgentsExists(c *partials.CommonInfo) (bool, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return false, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return false, err
	}

	if siteID == -1 {
		return m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Exist(context.Background())
	} else {
		return m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Exist(context.Background())
	}
}

func (m *Model) DeleteAllAgents(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Agent.Delete().Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	} else {
		return m.Client.Agent.Delete().Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	}
}

func (m *Model) SaveAgentUpdateInfo(agentId, status, description, version string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return m.Client.Agent.UpdateOneID(agentId).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			SetUpdateTaskStatus(status).
			SetUpdateTaskDescription(description).
			SetUpdateTaskExecution(time.Time{}).
			SetUpdateTaskVersion(version).
			SetUpdateTaskResult("").Exec(context.Background())
	} else {
		return m.Client.Agent.UpdateOneID(agentId).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			SetUpdateTaskStatus(status).
			SetUpdateTaskDescription(description).
			SetUpdateTaskExecution(time.Time{}).
			SetUpdateTaskVersion(version).
			SetUpdateTaskResult("").Exec(context.Background())
	}
}

func (m *Model) GetUpdateAgentsByPage(p partials.PaginationAndSort, f filters.UpdateAgentsFilter, c *partials.CommonInfo) ([]*ent.Agent, error) {
	var err error
	var agents []*ent.Agent

	query := m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).WithTags().WithRelease().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = query.Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = query.Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	// Apply filters
	applyUpdateAgentsFilters(query, f)

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldHostname)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldHostname)).All(context.Background())
		}
	case "version":
		if p.SortOrder == "asc" {
			agents, err = query.Order(agent.ByReleaseField(release.FieldVersion, sql.OrderAsc())).All(context.Background())
		} else {
			agents, err = query.Order(agent.ByReleaseField(release.FieldVersion, sql.OrderDesc())).All(context.Background())
		}
	case "taskStatus":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldUpdateTaskStatus)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskStatus)).All(context.Background())
		}
	case "taskDescription":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldUpdateTaskDescription)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskDescription)).All(context.Background())
		}
	case "taskLastExecution":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldUpdateTaskExecution)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskExecution)).All(context.Background())
		}
	case "taskResult":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldUpdateTaskResult)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskResult)).All(context.Background())
		}
	default:
		agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskExecution)).All(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (m *Model) CountAllUpdateAgents(f filters.UpdateAgentsFilter, c *partials.CommonInfo) (int, error) {
	var query *ent.AgentQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		query = m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	applyUpdateAgentsFilters(query, f)

	count, err := query.Count(context.Background())
	return count, err
}

func (m *Model) GetAllUpdateAgents(f filters.UpdateAgentsFilter, c *partials.CommonInfo) ([]*ent.Agent, error) {
	var query *ent.AgentQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}
	// Apply filters
	applyUpdateAgentsFilters(query, f)

	agents, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (m *Model) SaveAgentSettings(agentID string, settings openuem_nats.AgentSetting, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		return m.Client.Agent.UpdateOneID(agentID).Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).SetDebugMode(settings.DebugMode).SetSftpPort(settings.SFTPPort).SetSftpService(settings.SFTPService).SetRemoteAssistance(settings.RemoteAssistance).SetVncProxyPort(settings.VNCProxyPort).SetSettingsModified(time.Now()).Save(context.Background())
	} else {
		return m.Client.Agent.UpdateOneID(agentID).Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).SetDebugMode(settings.DebugMode).SetSftpPort(settings.SFTPPort).SetSftpService(settings.SFTPService).SetRemoteAssistance(settings.RemoteAssistance).SetVncProxyPort(settings.VNCProxyPort).SetSettingsModified(time.Now()).Save(context.Background())
	}
}

func applyUpdateAgentsFilters(query *ent.AgentQuery, f filters.UpdateAgentsFilter) {
	if len(f.Hostname) > 0 {
		query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if len(f.Releases) > 0 {
		query.Where(agent.HasReleaseWith(release.VersionIn(f.Releases...)))
	}

	if len(f.Tags) > 0 {
		predicates := []predicate.Agent{}
		for _, id := range f.Tags {
			predicates = append(predicates, agent.HasTagsWith(tag.ID(id)))
		}
		if len(predicates) > 0 {
			query.Where(agent.And(predicates...))
		}
	}

	if len(f.TaskStatus) > 0 {
		query.Where(agent.UpdateTaskStatusIn(f.TaskStatus...))
	}

	if len(f.TaskResult) > 0 {
		query.Where(agent.UpdateTaskResultContainsFold(f.TaskResult))
	}

	if len(f.TaskLastExecutionFrom) > 0 {
		from, err := time.Parse("2006-01-02", f.TaskLastExecutionFrom)
		if err == nil {
			query.Where(agent.UpdateTaskExecutionGTE(from))
		}
	}

	if len(f.TaskLastExecutionTo) > 0 {
		to, err := time.Parse("2006-01-02", f.TaskLastExecutionTo)
		if err == nil {
			query.Where(agent.UpdateTaskExecutionLTE(to))
		}
	}
}

func (m *Model) UpdateRemoteAssistanceToAllAgents(status bool, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		if _, err := m.Client.Agent.Update().Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).SetRemoteAssistance(status).Save(context.Background()); err != nil {
			return err
		}
	} else {
		if _, err := m.Client.Agent.Update().Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).SetRemoteAssistance(status).Save(context.Background()); err != nil {
			return err
		}

	}
	return nil
}

func (m *Model) UpdateSFTPServiceToAllAgents(status bool, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		if _, err := m.Client.Agent.Update().Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).SetSftpService(status).Save(context.Background()); err != nil {
			return err
		}
	} else {
		if _, err := m.Client.Agent.Update().Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).SetSftpService(status).Save(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) AssociateDefaultSiteToAgents(site *ent.Site) error {
	return m.Client.Agent.Update().Where(agent.Not(agent.HasSite())).AddSite(site).Exec(context.Background())
}
