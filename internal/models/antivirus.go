package models

import (
	"context"
	"slices"
	"strconv"

	"entgo.io/ent/dialect/sql"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/antivirus"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type Antivirus struct {
	ID        string
	Nickname  string
	OS        string
	Name      string
	IsActive  bool `sql:"is_active"`
	IsUpdated bool `sql:"is_updated"`
	SiteID    int
}

func mainAntivirusQuery(s *sql.Selector, p partials.PaginationAndSort) {
	// Info from agents waiting for admission won't be shown
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldNickname, agent.FieldOs, antivirus.FieldName, antivirus.FieldIsActive, antivirus.FieldIsUpdated).
		LeftJoin(sql.Table(antivirus.Table)).
		On(agent.FieldID, antivirus.OwnerColumn).
		Where(sql.And(sql.NEQ(agent.FieldAgentStatus, agent.AgentStatusWaitingForAdmission)))

	if p.PageSize != 0 {
		s.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
	}
}

func (m *Model) CountAllAntiviri(f filters.AgentFilter, c *partials.CommonInfo) (int, error) {
	var query *ent.AgentQuery
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	// Info from agents waiting for admission won't be shown

	if siteID == -1 {
		query = m.Client.Agent.Query().
			Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))

	} else {
		query = m.Client.Agent.Query().
			Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	applyEDRFilters(query, &f)

	return query.Count(context.Background())
}

func (m *Model) GetAntiviriByPage(p partials.PaginationAndSort, f filters.AgentFilter, c *partials.CommonInfo) ([]Antivirus, error) {
	var query *ent.AgentQuery
	var antiviri []Antivirus
	var err error

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Agent.Query().
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = m.Client.Agent.Query().
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	applyEDRFilters(query, &f)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "nickname"
		p.SortOrder = "desc"
	}

	switch p.SortBy {
	case "nickname":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldNickname))
			}).Scan(context.Background(), &antiviri)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldNickname))
			}).Scan(context.Background(), &antiviri)
		}
	case "agentOS":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldOs))
			}).Scan(context.Background(), &antiviri)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldOs))
			}).Scan(context.Background(), &antiviri)
		}
	case "antivirusName":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Asc(antivirus.FieldName))
			}).Scan(context.Background(), &antiviri)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Desc(antivirus.FieldName))
			}).Scan(context.Background(), &antiviri)
		}
	case "antivirusEnabled":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Asc(antivirus.FieldIsActive))
			}).Scan(context.Background(), &antiviri)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Desc(antivirus.FieldIsActive))
			}).Scan(context.Background(), &antiviri)
		}
	case "antivirusUpdated":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Asc(antivirus.FieldIsUpdated))
			}).Scan(context.Background(), &antiviri)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Desc(antivirus.FieldIsUpdated))
			}).Scan(context.Background(), &antiviri)
		}
	}

	if err != nil {
		return nil, err
	}

	// Add site ids
	sortedAgentIDs := []string{}
	for _, computer := range antiviri {
		sortedAgentIDs = append(sortedAgentIDs, computer.ID)
	}
	agents, err := m.Client.Agent.Query().WithSite().Where(agent.IDIn(sortedAgentIDs...)).All(context.Background())
	if err != nil {
		return nil, err
	}

	// Add site id to each computer in order
	for i, computer := range antiviri {
		for _, agent := range agents {
			if computer.ID == agent.ID {
				if len(agent.Edges.Site) == 1 {
					antiviri[i].SiteID = agent.Edges.Site[0].ID
				} else {
					antiviri[i].SiteID = -1
				}
				break
			}
		}
	}

	return antiviri, nil
}

func (m *Model) GetDetectedAntiviri(c *partials.CommonInfo, f filters.AgentFilter) ([]string, error) {
	var query *ent.AntivirusQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Antivirus.Query().Unique(true).
			Where(antivirus.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission),
				agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.Antivirus.Query().Unique(true).
			Where(antivirus.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission),
				agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

	if len(f.Nickname) > 0 {
		query.Where(antivirus.HasOwnerWith(agent.NicknameContainsFold(f.Nickname)))
	}

	if len(f.AgentOSVersions) > 0 {
		query.Where(antivirus.HasOwnerWith(agent.OsIn(f.AgentOSVersions...)))
	}

	if len(f.AntivirusNameOptions) > 0 {
		query.Where(antivirus.HasOwnerWith(agent.HasAntivirusWith(antivirus.NameIn(f.AntivirusNameOptions...))))
	}

	if len(f.AntivirusEnabledOptions) > 0 {
		if len(f.AntivirusEnabledOptions) == 1 && f.AntivirusEnabledOptions[0] == "Enabled" {
			query.Where(antivirus.HasOwnerWith(agent.HasAntivirusWith(antivirus.IsActive(true))))
		}

		if len(f.AntivirusEnabledOptions) == 1 && f.AntivirusEnabledOptions[0] == "Disabled" {
			query.Where(antivirus.HasOwnerWith(agent.HasAntivirusWith(antivirus.IsActive(false))))
		}
	}

	if len(f.AntivirusUpdatedOptions) > 0 {
		if len(f.AntivirusUpdatedOptions) == 1 && f.AntivirusUpdatedOptions[0] == "UpdatedYes" {
			query.Where(antivirus.HasOwnerWith(agent.HasAntivirusWith(antivirus.IsUpdated(true))))
		}

		if len(f.AntivirusUpdatedOptions) == 1 && f.AntivirusUpdatedOptions[0] == "UpdatedNo" {
			query.Where(antivirus.HasOwnerWith(agent.HasAntivirusWith(antivirus.IsUpdated(false))))
		}
	}

	if len(f.Search) > 0 {
		query.Where(antivirus.HasOwnerWith(agent.Or(
			agent.NicknameContainsFold(f.Search),
			agent.OsContainsFold(f.Search),
			agent.HasAntivirusWith(antivirus.NameContainsFold(f.Search)))))
	}

	// Remove results where antivirus name is empty
	query.Where(antivirus.NameNEQ(""))

	return query.Select(antivirus.FieldName).Strings(context.Background())
}

func applyEDRFilters(query *ent.AgentQuery, f *filters.AgentFilter) {
	if len(f.Nickname) > 0 {
		query.Where(agent.NicknameContainsFold(f.Nickname))
	}

	if len(f.AgentOSVersions) > 0 {
		query.Where(agent.OsIn(f.AgentOSVersions...))
	}

	if len(f.AntivirusNameOptions) > 0 {
		query.Where(agent.HasAntivirusWith(antivirus.NameIn(f.AntivirusNameOptions...)))
	}

	if len(f.AntivirusEnabledOptions) > 0 {
		if len(f.AntivirusEnabledOptions) == 1 && f.AntivirusEnabledOptions[0] == "Enabled" {
			query.Where(agent.HasAntivirusWith(antivirus.IsActive(true)))
		}

		if len(f.AntivirusEnabledOptions) == 1 && f.AntivirusEnabledOptions[0] == "Disabled" {
			query.Where(agent.HasAntivirusWith(antivirus.IsActive(false)))
		}
	}

	if len(f.AntivirusUpdatedOptions) > 0 {
		if len(f.AntivirusUpdatedOptions) == 1 && f.AntivirusUpdatedOptions[0] == "UpdatedYes" {
			query.Where(agent.HasAntivirusWith(antivirus.IsUpdated(true)))
		}

		if len(f.AntivirusUpdatedOptions) == 1 && f.AntivirusUpdatedOptions[0] == "UpdatedNo" {
			query.Where(agent.HasAntivirusWith(antivirus.IsUpdated(false)))
		}
	}

	if len(f.Search) > 0 {
		query.Where(agent.Or(
			agent.NicknameContainsFold(f.Search),
			agent.OsContainsFold(f.Search),
			agent.HasAntivirusWith(antivirus.NameContainsFold(f.Search))))
	}

	// Remove results where antivirus name is empty
	query.Where(agent.HasAntivirusWith(antivirus.NameNEQ("")))
}

func (m *Model) GetEDRUpdateStatusOptions(c *partials.CommonInfo, f *filters.AgentFilter) ([]string, error) {
	var query *ent.AntivirusQuery
	var err error

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Antivirus.Query().Unique(true).
			Where(antivirus.HasOwnerWith(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.Antivirus.Query().Unique(true).
			Where(antivirus.HasOwnerWith(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

	applyEDROptionsFilters(query, f)

	values, err := query.Select(antivirus.FieldIsUpdated).Strings(context.Background())
	if err != nil {
		return nil, err
	}

	options := []string{}

	if slices.Contains(values, "true") {
		options = append(options, "UpdatedYes")
	}

	if slices.Contains(values, "false") {
		options = append(options, "UpdatedNo")
	}

	return options, nil
}

func (m *Model) GetEDREnabledStatusOptions(c *partials.CommonInfo, f *filters.AgentFilter) ([]string, error) {
	var query *ent.AntivirusQuery
	var err error

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Antivirus.Query().Unique(true).
			Where(antivirus.HasOwnerWith(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.Antivirus.Query().Unique(true).
			Where(antivirus.HasOwnerWith(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

	applyEDROptionsFilters(query, f)

	values, err := query.Select(antivirus.FieldIsActive).Strings(context.Background())
	if err != nil {
		return nil, err
	}

	options := []string{}

	if slices.Contains(values, "true") {
		options = append(options, "Enabled")
	}

	if slices.Contains(values, "false") {
		options = append(options, "Disabled")
	}

	return options, nil
}

func (m *Model) GetEDRNamesOptions(c *partials.CommonInfo, f *filters.AgentFilter) ([]string, error) {
	var query *ent.AntivirusQuery
	var err error

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Antivirus.Query().Unique(true).
			Where(antivirus.HasOwnerWith(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.Antivirus.Query().Unique(true).
			Where(antivirus.HasOwnerWith(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

	applyEDROptionsFilters(query, f)

	values, err := query.Select(antivirus.FieldName).Strings(context.Background())
	if err != nil {
		return nil, err
	}

	return values, nil
}

func applyEDROptionsFilters(query *ent.AntivirusQuery, f *filters.AgentFilter) {
	if len(f.Nickname) > 0 {
		query.Where(antivirus.HasOwnerWith(agent.NicknameContainsFold(f.Nickname)))
	}

	if len(f.AgentOSVersions) > 0 {
		query.Where(antivirus.HasOwnerWith(agent.OsIn(f.AgentOSVersions...)))
	}

	if len(f.AntivirusNameOptions) > 0 {
		query.Where(antivirus.NameIn(f.AntivirusNameOptions...))
	}

	if len(f.AntivirusEnabledOptions) > 0 {
		if len(f.AntivirusEnabledOptions) == 1 && f.AntivirusEnabledOptions[0] == "Enabled" {
			query.Where(antivirus.IsActive(true))
		}

		if len(f.AntivirusEnabledOptions) == 1 && f.AntivirusEnabledOptions[0] == "Disabled" {
			query.Where(antivirus.IsActive(false))
		}
	}

	if len(f.AntivirusUpdatedOptions) > 0 {
		if len(f.AntivirusUpdatedOptions) == 1 && f.AntivirusUpdatedOptions[0] == "UpdatedYes" {
			query.Where(antivirus.IsUpdated(true))
		}

		if len(f.AntivirusUpdatedOptions) == 1 && f.AntivirusUpdatedOptions[0] == "UpdatedNo" {
			query.Where(antivirus.IsUpdated(false))
		}
	}

	if len(f.Search) > 0 {
		query.Where(antivirus.Or(
			antivirus.HasOwnerWith(agent.Or(agent.NicknameContainsFold(f.Search), agent.OsContainsFold(f.Search))),
			antivirus.NameContainsFold(f.Search)))
	}

	// Remove results where antivirus name is empty
	query.Where(antivirus.NameNEQ(""))
}
