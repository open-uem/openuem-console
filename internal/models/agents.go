package models

import (
	"context"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/operatingsystem"
	"github.com/doncicuto/openuem_ent/predicate"
	"github.com/doncicuto/openuem_ent/systemupdate"
	"github.com/doncicuto/openuem_ent/tag"
)

type Agent struct {
	OS      string
	Version string
	Status  string
	Count   int
}

func (m *Model) GetAllAgents() ([]*ent.Agent, error) {
	agents, err := m.Client.Agent.Query().All(context.Background())
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (m *Model) GetAgentsByPage(p partials.PaginationAndSort, f filters.AgentFilter) ([]*ent.Agent, error) {
	var err error
	var apps []*ent.Agent

	query := m.Client.Agent.Query().WithTags().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	// Apply filters
	applyAgentFilters(query, f)

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			apps, err = query.Order(ent.Asc(agent.FieldHostname)).All(context.Background())
		} else {
			apps, err = query.Order(ent.Desc(agent.FieldHostname)).All(context.Background())
		}
	case "os":
		if p.SortOrder == "asc" {
			apps, err = query.Order(ent.Asc(agent.FieldOs)).All(context.Background())
		} else {
			apps, err = query.Order(ent.Desc(agent.FieldOs)).All(context.Background())
		}
	case "version":
		if p.SortOrder == "asc" {
			apps, err = query.Order(ent.Asc(agent.FieldVersion)).All(context.Background())
		} else {
			apps, err = query.Order(ent.Desc(agent.FieldVersion)).All(context.Background())
		}
	case "last_contact":
		if p.SortOrder == "asc" {
			apps, err = query.Order(ent.Asc(agent.FieldLastContact)).All(context.Background())
		} else {
			apps, err = query.Order(ent.Desc(agent.FieldLastContact)).All(context.Background())
		}
	case "enabled":
		if p.SortOrder == "asc" {
			apps, err = query.Order(ent.Asc(agent.FieldEnabled)).All(context.Background())
		} else {
			apps, err = query.Order(ent.Desc(agent.FieldEnabled)).All(context.Background())
		}
	case "ip_address":
		if p.SortOrder == "asc" {
			apps, err = query.Order(ent.Asc(agent.FieldIP)).All(context.Background())
		} else {
			apps, err = query.Order(ent.Desc(agent.FieldIP)).All(context.Background())
		}
	default:
		apps, err = query.Order(ent.Desc(agent.FieldLastContact)).All(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (m *Model) GetAgentById(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, err
}

func (m *Model) CountAgentsByOS() ([]Agent, error) {
	agents := []Agent{}
	err := m.Client.Agent.Query().Modify(func(s *sql.Selector) {
		s.Select(agent.FieldOs, sql.As(sql.Count("os"), "count")).GroupBy("os").OrderBy("count")
	}).Scan(context.Background(), &agents)
	if err != nil {
		return nil, err
	}
	return agents, err
}

func (m *Model) CountAllAgents(f filters.AgentFilter) (int, error) {
	query := m.Client.Agent.Query()

	applyAgentFilters(query, f)

	count, err := query.Count(context.Background())
	return count, err
}

func applyAgentFilters(query *ent.AgentQuery, f filters.AgentFilter) {
	if len(f.Hostname) > 0 {
		query = query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if f.EnabledAgents && !f.DisabledAgents {
		query = query.Where(agent.Enabled(true))
	}

	if f.DisabledAgents && !f.EnabledAgents {
		query = query.Where(agent.Enabled(false))
	}

	if f.WindowsAgents || f.LinuxAgents || f.MacAgents {
		agentSystems := []string{}

		if f.WindowsAgents {
			agentSystems = append(agentSystems, "windows")
		}
		if f.LinuxAgents {
			agentSystems = append(agentSystems, "linux")
		}
		if f.MacAgents {
			agentSystems = append(agentSystems, "mac")
		}

		query = query.Where(agent.OsIn(agentSystems...))
	}

	if len(f.Tags) > 0 {
		predicates := []predicate.Agent{}
		for _, id := range f.Tags {
			predicates = append(predicates, agent.HasTagsWith(tag.ID(id)))
		}
		if len(predicates) > 0 {
			query = query.Where(agent.And(predicates...))
		}
	}
}

func (m *Model) CountAgentsByOSVersion() ([]Agent, error) {
	agents := []Agent{}
	err := m.Client.OperatingSystem.Query().GroupBy(operatingsystem.FieldVersion).Aggregate(ent.Count()).Scan(context.Background(), &agents)
	if err != nil {
		return nil, err
	}
	return agents, err
}

func (m *Model) CountAgentsReportedLast24h() (int, error) {
	count, err := m.Client.Agent.Query().Where(agent.LastContactGTE(time.Now().AddDate(0, 0, -1))).Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, err
}

func (m *Model) CountAgentsByWindowsUpdateStatus() ([]Agent, error) {
	agents := []Agent{}
	err := m.Client.SystemUpdate.Query().GroupBy(systemupdate.FieldStatus).Aggregate(ent.Count()).Scan(context.Background(), &agents)
	if err != nil {
		return nil, err
	}
	return agents, err
}

func (m *Model) DeleteAgent(agentId string) error {
	err := m.Client.Agent.DeleteOneID(agentId).Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) EnableAgent(agentId string) error {
	_, err := m.Client.Agent.UpdateOneID(agentId).SetEnabled(true).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) DisableAgent(agentId string) error {
	_, err := m.Client.Agent.UpdateOneID(agentId).SetEnabled(false).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) AddTagToAgent(agentId, tagId string) error {
	id, err := strconv.Atoi(tagId)
	if err != nil {
		return err
	}
	return m.Client.Agent.UpdateOneID(agentId).AddTagIDs(id).Exec(context.Background())
}

func (m *Model) RemoveTagFromAgent(agentId, tagId string) error {
	id, err := strconv.Atoi(tagId)
	if err != nil {
		return err
	}
	return m.Client.Agent.UpdateOneID(agentId).RemoveTagIDs(id).Exec(context.Background())
}
