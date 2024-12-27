package models

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/antivirus"
)

type Antivirus struct {
	ID        string
	Hostname  string
	OS        string
	Name      string
	IsActive  bool `sql:"is_active"`
	IsUpdated bool `sql:"is_updated"`
}

func mainAntivirusQuery(s *sql.Selector, p partials.PaginationAndSort) {
	// Info from agents waiting for admission won't be shown
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldHostname, agent.FieldOs, antivirus.FieldName, antivirus.FieldIsActive, antivirus.FieldIsUpdated).
		LeftJoin(sql.Table(antivirus.Table)).
		On(agent.FieldID, antivirus.OwnerColumn).
		Where(sql.And(sql.NEQ(agent.FieldAgentStatus, agent.AgentStatusWaitingForAdmission))).
		Limit(p.PageSize).
		Offset((p.CurrentPage - 1) * p.PageSize)
}

func (m *Model) CountAllAntiviri(f filters.AntivirusFilter) (int, error) {
	// Info from agents waiting for admission won't be shown
	query := m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))

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

func (m *Model) GetDetectedAntiviri() ([]string, error) {
	return m.Client.Antivirus.Query().Unique(true).Where(antivirus.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))).Select(antivirus.FieldName).Strings(context.Background())
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
