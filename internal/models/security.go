package models

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/antivirus"
	"github.com/doncicuto/openuem_ent/systemupdate"
	"github.com/doncicuto/openuem_ent/update"
)

type Antivirus struct {
	ID        string
	Hostname  string
	OS        string
	Name      string
	IsActive  bool `sql:"is_active"`
	IsUpdated bool `sql:"is_updated"`
}

type SystemUpdate struct {
	ID             string
	Hostname       string
	OS             string
	Status         string
	LastInstall    time.Time `sql:"last_install"`
	LastSearch     time.Time `sql:"last_search"`
	PendingUpdates bool      `sql:"pending_updates"`
}

func mainAntivirusQuery(s *sql.Selector, p partials.PaginationAndSort) {
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldHostname, agent.FieldOs, antivirus.FieldName, antivirus.FieldIsActive, antivirus.FieldIsUpdated).
		LeftJoin(sql.Table(antivirus.Table)).
		On(agent.FieldID, antivirus.OwnerColumn).
		Limit(p.PageSize).
		Offset((p.CurrentPage - 1) * p.PageSize)
}

func mainUpdatesQuery(s *sql.Selector, p partials.PaginationAndSort) {
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldHostname, agent.FieldOs, systemupdate.FieldStatus, systemupdate.FieldLastInstall, systemupdate.FieldLastSearch, systemupdate.FieldPendingUpdates).
		LeftJoin(sql.Table(systemupdate.Table)).
		On(agent.FieldID, systemupdate.OwnerColumn).
		Limit(p.PageSize).
		Offset((p.CurrentPage - 1) * p.PageSize)
}

func (m *Model) CountAllAntiviri(f filters.AntivirusFilter) (int, error) {
	query := m.Client.Agent.Query()

	applyAntiviriFilters(query, f)

	return query.Count(context.Background())
}

func (m *Model) GetAntiviriByPage(p partials.PaginationAndSort, f filters.AntivirusFilter) ([]Antivirus, error) {
	var antiviri []Antivirus
	var err error

	query := m.Client.Agent.Query()

	applyAntiviriFilters(query, f)

	// Default sort
	if p.SortBy == "" {
		p.SortBy = "hostname"
		p.SortOrder = "desc"
	}

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldHostname))
			}).Scan(context.Background(), &antiviri)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainAntivirusQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldHostname))
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

	return antiviri, nil
}

func applyAntiviriFilters(query *ent.AgentQuery, f filters.AntivirusFilter) {
	if len(f.Hostname) > 0 {
		query = query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if len(f.AgentOSVersions) > 0 {
		query = query.Where(agent.OsIn(f.AgentOSVersions...))
	}

	if len(f.AntivirusNameOptions) > 0 {
		query = query.Where(agent.HasAntivirusWith(antivirus.NameIn(f.AntivirusNameOptions...)))
	}

	if len(f.AntivirusEnabledOptions) > 0 {
		if len(f.AntivirusEnabledOptions) == 1 && f.AntivirusEnabledOptions[0] == "Enabled" {
			query = query.Where(agent.HasAntivirusWith(antivirus.IsActive(true)))
		}

		if len(f.AntivirusEnabledOptions) == 1 && f.AntivirusEnabledOptions[0] == "Disabled" {
			query = query.Where(agent.HasAntivirusWith(antivirus.IsActive(false)))
		}
	}

	if len(f.AntivirusUpdatedOptions) > 0 {
		if len(f.AntivirusUpdatedOptions) == 1 && f.AntivirusUpdatedOptions[0] == "UpdatedYes" {
			query = query.Where(agent.HasAntivirusWith(antivirus.IsUpdated(true)))
		}

		if len(f.AntivirusUpdatedOptions) == 1 && f.AntivirusUpdatedOptions[0] == "UpdatedNo" {
			query = query.Where(agent.HasAntivirusWith(antivirus.IsUpdated(false)))
		}
	}
}

func (m *Model) CountAllSystemUpdates(f filters.SystemUpdatesFilter) (int, error) {
	query := m.Client.Agent.Query()

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
				s.OrderBy(sql.Asc(systemupdate.FieldStatus))
			}).Scan(context.Background(), &systemUpdates)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainUpdatesQuery(s, p)
				s.OrderBy(sql.Desc(systemupdate.FieldStatus))
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
		query = query.Where(agent.HasSystemupdateWith(systemupdate.StatusIn(f.UpdateStatus...)))
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

func (m *Model) CountLatestUpdates(agentId string) (int, error) {
	return m.Client.Update.Query().Where(update.HasOwnerWith(agent.ID(agentId))).Count(context.Background())
}

func (m *Model) GetLatestUpdates(agentId string, p partials.PaginationAndSort) ([]*ent.Update, error) {
	query := m.Client.Update.Query().Where(update.HasOwnerWith(agent.ID(agentId)))

	switch p.SortBy {
	case "title":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(update.FieldTitle))
		} else {
			query = query.Order(ent.Desc(update.FieldTitle))
		}
	case "date":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(update.FieldDate))
		} else {
			query = query.Order(ent.Desc(update.FieldDate))
		}
	}

	updates, err := query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
	if err != nil {
		return nil, err
	}

	return updates, nil
}

func (m *Model) GetDetectedAntiviri() ([]string, error) {
	return m.Client.Antivirus.Query().Unique(true).Select(antivirus.FieldName).Strings(context.Background())
}
