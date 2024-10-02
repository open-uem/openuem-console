package models

import (
	"context"

	"entgo.io/ent/dialect/sql"
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

func (m *Model) CountAllApps() (int, error) {
	var apps []App

	err := m.Client.App.Query().GroupBy(app.FieldName).Scan(context.Background(), &apps)
	if err != nil {
		return 0, err
	}
	return len(apps), err
}

func (m *Model) GetAgentAppsByPage(agentId string, currentPage, nAppsPerPage int) ([]*ent.App, error) {
	apps, err := m.Client.App.Query().Where(app.HasOwnerWith(agent.ID(agentId))).Limit(nAppsPerPage).Offset((currentPage - 1) * nAppsPerPage).Order(ent.Asc(app.FieldName)).All(context.Background())
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func mainAppsByPageSQL(s *sql.Selector, currentPage, nAppsPerPage int) {
	s.Select(app.FieldName, app.FieldPublisher, "count(*) AS count").GroupBy(app.FieldName, app.FieldPublisher).Limit(nAppsPerPage).Offset((currentPage - 1) * nAppsPerPage)
}

func (m *Model) GetAppsByPage(currentPage, nAppsPerPage int, sortBy, sortOrder string) ([]App, error) {
	var apps []App
	var err error

	switch sortBy {
	case "name":
		if sortOrder == "asc" {
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, currentPage, nAppsPerPage)
				s.OrderBy(sql.Asc(app.FieldName))
			}).Scan(context.Background(), &apps)
		} else {
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, currentPage, nAppsPerPage)
				s.OrderBy(sql.Desc(app.FieldName))
			}).Scan(context.Background(), &apps)
		}
	case "publisher":
		if sortOrder == "asc" {
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, currentPage, nAppsPerPage)
				s.OrderBy(sql.Asc(app.FieldPublisher))
			}).Scan(context.Background(), &apps)
		} else {
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, currentPage, nAppsPerPage)
				s.OrderBy(sql.Desc(app.FieldPublisher))
			}).Scan(context.Background(), &apps)
		}
	case "installations":
		if sortOrder == "asc" {
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, currentPage, nAppsPerPage)
				s.OrderBy(sql.Asc("count"))
			}).Scan(context.Background(), &apps)
		} else {
			err = m.Client.App.Query().Modify(func(s *sql.Selector) {
				mainAppsByPageSQL(s, currentPage, nAppsPerPage)
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
