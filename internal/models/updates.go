package models

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/systemupdate"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type SystemUpdate struct {
	ID                 string
	Hostname           string
	OS                 string
	SystemUpdateStatus string    `sql:"system_update_status"`
	LastInstall        time.Time `sql:"last_install"`
	LastSearch         time.Time `sql:"last_search"`
	PendingUpdates     bool      `sql:"pending_updates"`
}

func mainUpdatesQuery(s *sql.Selector, p partials.PaginationAndSort) {
	// Info from agents waiting for admission won't be shown
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldHostname, agent.FieldOs, systemupdate.FieldSystemUpdateStatus, systemupdate.FieldLastInstall, systemupdate.FieldLastSearch, systemupdate.FieldPendingUpdates).
		LeftJoin(sql.Table(systemupdate.Table)).
		On(agent.FieldID, systemupdate.OwnerColumn).
		Where(sql.And(sql.NEQ(agent.FieldAgentStatus, agent.AgentStatusWaitingForAdmission))).
		Limit(p.PageSize).
		Offset((p.CurrentPage - 1) * p.PageSize)
}

func (m *Model) CountAllSystemUpdates(f filters.SystemUpdatesFilter) (int, error) {
	query := m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))

	applySystemUpdatesFilters(query, f)

	return query.Count(context.Background())
}

func (m *Model) GetSystemUpdatesByPage(p partials.PaginationAndSort, f filters.SystemUpdatesFilter) ([]SystemUpdate, error) {
	var systemUpdates []SystemUpdate
	var err error

	query := m.Client.Agent.Query()

	applySystemUpdatesFilters(query, f)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "hostname"
		p.SortOrder = "desc"
	}

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldHostname))
			}).Scan(context.Background(), &systemUpdates)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldHostname))
			}).Scan(context.Background(), &systemUpdates)
		}
	case "agentOS":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldOs))
			}).Scan(context.Background(), &systemUpdates)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldOs))
			}).Scan(context.Background(), &systemUpdates)
		}
	case "updateStatus":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Asc(systemupdate.FieldSystemUpdateStatus))
			}).Scan(context.Background(), &systemUpdates)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Desc(systemupdate.FieldSystemUpdateStatus))
			}).Scan(context.Background(), &systemUpdates)
		}
	case "lastSearch":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Asc(systemupdate.FieldLastSearch))
			}).Scan(context.Background(), &systemUpdates)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Desc(systemupdate.FieldLastSearch))
			}).Scan(context.Background(), &systemUpdates)
		}
	case "lastInstall":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Asc(systemupdate.FieldLastInstall))
			}).Scan(context.Background(), &systemUpdates)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Desc(systemupdate.FieldLastInstall))
			}).Scan(context.Background(), &systemUpdates)
		}
	case "pendingUpdates":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Asc(systemupdate.FieldPendingUpdates))
			}).Scan(context.Background(), &systemUpdates)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Desc(systemupdate.FieldPendingUpdates))
			}).Scan(context.Background(), &systemUpdates)
		}
	}

	if err != nil {
		return nil, err
	}

	return systemUpdates, nil
}

func applySystemUpdatesFilters(query *ent.AgentQuery, f filters.SystemUpdatesFilter) {
	if len(f.Hostname) > 0 {
		query = query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if len(f.AgentOSVersions) > 0 {
		query = query.Where(agent.OsIn(f.AgentOSVersions...))
	}

	if len(f.UpdateStatus) > 0 {
		query = query.Where(agent.HasSystemupdateWith(systemupdate.SystemUpdateStatusIn(f.UpdateStatus...)))
	}

	if len(f.LastSearchFrom) > 0 {
		from, err := time.Parse("2006-01-02", f.LastSearchFrom)
		if err == nil {
			query = query.Where(agent.HasSystemupdateWith(systemupdate.LastSearchGTE(from)))
		}
	}

	if len(f.LastSearchTo) > 0 {
		to, err := time.Parse("2006-01-02", f.LastSearchTo)
		if err == nil {
			query = query.Where(agent.HasSystemupdateWith(systemupdate.LastSearchLTE(to)))
		}
	}

	if len(f.LastInstallFrom) > 0 {
		from, err := time.Parse("2006-01-02", f.LastInstallFrom)
		if err == nil {
			query = query.Where(agent.HasSystemupdateWith(systemupdate.LastInstallGTE(from)))
		}
	}

	if len(f.LastInstallTo) > 0 {
		to, err := time.Parse("2006-01-02", f.LastInstallTo)
		if err == nil {
			query = query.Where(agent.HasSystemupdateWith(systemupdate.LastInstallLTE(to)))
		}
	}

	if len(f.PendingUpdateOptions) > 0 {
		if len(f.PendingUpdateOptions) == 1 && f.PendingUpdateOptions[0] == "Yes" {
			query = query.Where(agent.HasSystemupdateWith(systemupdate.PendingUpdates(true)))
		}

		if len(f.PendingUpdateOptions) == 1 && f.PendingUpdateOptions[0] == "No" {
			query = query.Where(agent.HasSystemupdateWith(systemupdate.PendingUpdates(false)))
		}
	}
}
