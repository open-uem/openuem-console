package models

import (
	"context"
	"strconv"

	"entgo.io/ent/dialect/sql"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/app"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type App struct {
	ID        int
	Source    string
	Name      string
	Publisher string
	Count     int
}

func (m *Model) CountAgentApps(agentId string, f filters.ApplicationsFilter, c *partials.CommonInfo) (int, error) {
	var query *ent.AppQuery

	// Info from agents waiting for admission won't be shown
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		query = m.Client.App.Query().Where(app.HasOwnerWith(agent.ID(agentId), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.App.Query().Where(app.HasOwnerWith(agent.ID(agentId), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

	applyAppsFilters(query, f)

	count, err := query.Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, err
}

func (m *Model) CountAllApps(f filters.ApplicationsFilter, c *partials.CommonInfo) (int, error) {
	var apps []App
	var query *ent.AppQuery

	// Info from agents waiting for admission won't be shown
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		query = m.Client.App.Query().Where(app.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.App.Query().Where(app.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

	applyAppsFilters(query, f)

	if err := query.GroupBy(app.FieldName).Scan(context.Background(), &apps); err != nil {
		return 0, err
	}
	return len(apps), err
}

func (m *Model) GetAgentAppsByPage(agentId string, p partials.PaginationAndSort, f filters.ApplicationsFilter, c *partials.CommonInfo) ([]*ent.App, error) {
	var query *ent.AppQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		// Info from agents waiting for admission won't be shown
		query = m.Client.App.Query().
			Where(app.HasOwnerWith(agent.ID(agentId), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
	} else {
		query = m.Client.App.Query().
			Where(app.HasOwnerWith(agent.ID(agentId), agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
	}
	applyAppsFilters(query, f)

	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(app.FieldName))
		} else {
			query = query.Order(ent.Desc(app.FieldName))
		}
	case "version":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(app.FieldVersion))
		} else {
			query = query.Order(ent.Desc(app.FieldVersion))
		}
	case "publisher":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(app.FieldPublisher))
		} else {
			query = query.Order(ent.Desc(app.FieldPublisher))
		}
	case "installation":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(app.FieldInstallDate))
		} else {
			query = query.Order(ent.Desc(app.FieldInstallDate))
		}
	}

	apps, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func mainAppsByPageSQL(s *sql.Selector, p partials.PaginationAndSort) {
	s.Select(app.FieldName, app.FieldPublisher, "count(*) AS count").GroupBy(app.FieldName, app.FieldPublisher)
	if p.PageSize != 0 {
		s.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
	}
}

func (m *Model) GetAppsByPage(p partials.PaginationAndSort, f filters.ApplicationsFilter, c *partials.CommonInfo) ([]App, error) {
	var apps []App
	var err error
	var query *ent.AppQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		// Info from agents waiting for admission won't be shown
		query = m.Client.App.Query().Where(app.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.App.Query().Where(app.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

	applyAppsFilters(query, f)

	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, p)
				s.OrderBy(sql.Asc(app.FieldName))
			}).Scan(context.Background(), &apps)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, p)
				s.OrderBy(sql.Desc(app.FieldName))
			}).Scan(context.Background(), &apps)
		}
	case "publisher":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, p)
				s.OrderBy(sql.Asc(app.FieldPublisher))
			}).Scan(context.Background(), &apps)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, p)
				s.OrderBy(sql.Desc(app.FieldPublisher))
			}).Scan(context.Background(), &apps)
		}
	case "installations":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, p)
				s.OrderBy(sql.Asc("count"))
			}).Scan(context.Background(), &apps)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, p)
				s.OrderBy(sql.Desc("count"))
			}).Scan(context.Background(), &apps)
		}
	}

	if err != nil {
		return nil, err
	}

	return apps, err
}

func (m *Model) GetTop10InstalledApps() ([]App, error) {
	var apps []App
	err := m.Client.App.Query().Modify(func(s *sql.Selector) {
		s.Select(app.FieldName, sql.As(sql.Count("*"), "count")).GroupBy(app.FieldName).OrderBy(sql.Desc("count")).Limit(10)
	}).Scan(context.Background(), &apps)
	if err != nil {
		return nil, err
	}
	return apps, err
}

func applyAppsFilters(query *ent.AppQuery, f filters.ApplicationsFilter) {
	if len(f.AppName) > 0 {
		query.Where(app.NameContainsFold(f.AppName))
	}

	if len(f.Vendor) > 0 {
		query.Where(app.PublisherContainsFold(f.Vendor))
	}

	if len(f.Version) > 0 {
		query.Where(app.VersionContainsFold(f.Version))
	}

	if len(f.Search) > 0 {
		query.Where(app.Or(app.NameContainsFold(f.Search), app.PublisherContainsFold(f.Search), app.VersionContainsFold(f.Search)))
	}
}
