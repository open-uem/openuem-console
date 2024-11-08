package models

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/app"
)

type App struct {
	ID        int
	Name      string
	Publisher string
	Count     int
}

func (m *Model) CountAgentApps(agentId string) (int, error) {
	count, err := m.Client.App.Query().Where(app.HasOwnerWith(agent.ID(agentId))).Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, err
}

func (m *Model) CountAllApps(f filters.ApplicationsFilter) (int, error) {
	var apps []App

	query := m.Client.App.Query()

	applyAppsFilters(query, f)

	err := query.GroupBy(app.FieldName).Scan(context.Background(), &apps)
	if err != nil {
		return 0, err
	}
	return len(apps), err
}

func (m *Model) GetAgentAppsByPage(agentId string, p partials.PaginationAndSort) ([]*ent.App, error) {

	query := m.Client.App.Query().Where(app.HasOwnerWith(agent.ID(agentId))).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

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
	s = s.Select(app.FieldName, app.FieldPublisher, "count(*) AS count").GroupBy(app.FieldName, app.FieldPublisher).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
}

func (m *Model) GetAppsByPage(p partials.PaginationAndSort, f filters.ApplicationsFilter) ([]App, error) {
	var apps []App
	var err error

	query := m.Client.App.Query()

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
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, p)
				s.OrderBy(sql.Desc(app.FieldPublisher))
			}).Scan(context.Background(), &apps)
		}
	case "installations":
		if p.SortOrder == "asc" {
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, p)
				s.OrderBy(sql.Asc("count"))
			}).Scan(context.Background(), &apps)
		} else {
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
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
}
