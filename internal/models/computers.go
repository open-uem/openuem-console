package models

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/app"
	"github.com/open-uem/ent/computer"
	"github.com/open-uem/ent/operatingsystem"
	"github.com/open-uem/ent/predicate"
	"github.com/open-uem/ent/tag"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type Computer struct {
	ID           string
	Hostname     string
	OS           string
	Version      string
	IP           string
	MAC          string
	Username     string
	Manufacturer string
	Model        string
	Serial       string
	IsRemote     bool      `sql:"is_remote"`
	LastContact  time.Time `sql:"last_contact"`
	Tags         []*ent.Tag
}

func (m *Model) CountAllComputers(f filters.AgentFilter) (int, error) {

	// Agents that haven't been admitted yet should not appear
	query := m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))

	// Apply filters
	applyComputerFilters(query, f)

	count, err := query.Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, err
}

func mainQuery(s *sql.Selector, p partials.PaginationAndSort) {
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldHostname, agent.FieldOs, "`t2`.`version`", agent.FieldIP, agent.FieldMAC, operatingsystem.FieldUsername, computer.FieldManufacturer, computer.FieldModel, computer.FieldSerial, agent.FieldIsRemote, agent.FieldLastContact).
		LeftJoin(sql.Table(computer.Table)).
		On(agent.FieldID, computer.OwnerColumn).
		LeftJoin(sql.Table(operatingsystem.Table)).
		On(agent.FieldID, operatingsystem.OwnerColumn)

	if p.PageSize != 0 {
		s.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
	}
}

/* func (m *Model) GetComputersByPage(p partials.PaginationAndSort, f filters.AgentFilter) ([]*ent.Agent, error) {

	// Apply sort using go as there's a bug in entgo: https://github.com/ent/ent/issues/3722
	// I get SQL state: 42803 errors due to try sortering using edge fields that are not
	// part of the groupby

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			query = query.Order(agent.ByHostname())
		} else {
			query = query.Order(agent.ByHostname(sql.OrderDesc()))
		}
	case "os":
		if p.SortOrder == "asc" {
			query = query.Order(agent.ByOs())
		} else {
			query = query.Order(agent.ByOs(sql.OrderDesc()))
		}
	case "version":
		if p.SortOrder == "asc" {
			query = query.Order(agent.ByOperatingsystemField(operatingsystem.FieldVersion))
		} else {
			query = query.Order(agent.ByOperatingsystemField(operatingsystem.FieldVersion, sql.OrderDesc()))
		}
	case "username":
		if p.SortOrder == "asc" {
			query = query.Order(agent.ByOperatingsystemField(operatingsystem.FieldUsername))
		} else {
			query = query.Order(agent.ByOperatingsystemField(operatingsystem.FieldUsername, sql.OrderDesc()))
		}
	case "manufacturer":
		if p.SortOrder == "asc" {
			query = query.Order(agent.ByComputerField(computer.FieldManufacturer))
		} else {
			query = query.Order(agent.ByComputerField(computer.FieldManufacturer, sql.OrderDesc()))
		}
	case "model":
		if p.SortOrder == "asc" {
			query = query.Order(agent.ByComputerField(computer.FieldModel))
		} else {
			query = query.Order(agent.ByComputerField(computer.FieldModel, sql.OrderDesc()))
		}
	}

	return agents, nil
}*/

func (m *Model) GetComputersByPage(p partials.PaginationAndSort, f filters.AgentFilter) ([]Computer, error) {
	var err error
	var computers []Computer

	// Agents that haven't been admitted yet should not appear
	query := m.Client.Agent.Query().Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))

	// Apply filters
	applyComputerFilters(query, f)

	// Apply sort
	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldHostname))
			}).Scan(context.Background(), &computers)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldHostname))
			}).Scan(context.Background(), &computers)
		}
	case "os":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldOs))
			}).Scan(context.Background(), &computers)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldOs))
			}).Scan(context.Background(), &computers)
		}
	case "version":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc("version"))
			}).Scan(context.Background(), &computers)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc("version"))
			}).Scan(context.Background(), &computers)
		}
	case "username":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(operatingsystem.FieldUsername))
			}).Scan(context.Background(), &computers)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(operatingsystem.FieldUsername))
			}).Scan(context.Background(), &computers)
		}
	case "manufacturer":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(computer.FieldManufacturer))
			}).Scan(context.Background(), &computers)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(computer.FieldManufacturer))
			}).Scan(context.Background(), &computers)
		}
	case "model":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(computer.FieldModel))
			}).Scan(context.Background(), &computers)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(computer.FieldModel))
			}).Scan(context.Background(), &computers)
		}
	case "remote":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldIsRemote))
			}).Scan(context.Background(), &computers)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldIsRemote))
			}).Scan(context.Background(), &computers)
		}
	default:
		err = query.Modify(func(s *sql.Selector) {
			mainQuery(s, p)
			s.OrderBy(sql.Desc(agent.FieldLastContact))
		}).Scan(context.Background(), &computers)
	}
	if err != nil {
		return nil, err
	}

	// Add tags
	sortedAgentIDs := []string{}
	for _, computer := range computers {
		sortedAgentIDs = append(sortedAgentIDs, computer.ID)
	}
	agents, err := m.Client.Agent.Query().WithTags().Where(agent.IDIn(sortedAgentIDs...)).All(context.Background())
	if err != nil {
		return nil, err
	}

	// Add tags to each computer in order
	for i, computer := range computers {
		for _, agent := range agents {
			if computer.ID == agent.ID {
				computers[i].Tags = agent.Edges.Tags
				break
			}
		}
	}

	return computers, nil
}

func applyComputerFilters(query *ent.AgentQuery, f filters.AgentFilter) {
	if len(f.Hostname) > 0 {
		query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if len(f.Username) > 0 {
		query.Where(agent.HasOperatingsystemWith(operatingsystem.UsernameContainsFold(f.Username)))
	}

	if len(f.AgentOSVersions) > 0 {
		query.Where(agent.OsIn(f.AgentOSVersions...))
	}

	if len(f.OSVersions) > 0 {
		query.Where(agent.HasOperatingsystemWith(operatingsystem.VersionIn(f.OSVersions...)))
	}

	if len(f.ComputerManufacturers) > 0 {
		query.Where(agent.HasComputerWith(computer.ManufacturerIn(f.ComputerManufacturers...)))
	}

	if len(f.ComputerModels) > 0 {
		query.Where(agent.HasComputerWith(computer.ModelIn(f.ComputerModels...)))
	}

	if len(f.WithApplication) > 0 {
		query.Where(agent.HasAppsWith(app.Name(f.WithApplication)))
	}

	if len(f.IsRemote) > 0 {
		if len(f.IsRemote) == 1 && f.IsRemote[0] == "Remote" {
			query.Where(agent.IsRemote(true))
		}

		if len(f.IsRemote) == 1 && f.IsRemote[0] == "Local" {
			query.Where(agent.IsRemote(false))
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
}

func (m *Model) GetAgentComputerInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithComputer().WithTags().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentOSInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithOperatingsystem().WithTags().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentNetworkAdaptersInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithNetworkadapters().WithTags().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentPrintersInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithPrinters().WithTags().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentLogicalDisksInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithLogicaldisks().WithTags().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentSharesInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithShares().WithTags().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentMonitorsInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithMonitors().WithTags().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) SaveNotes(agentId string, notes string) error {
	return m.Client.Agent.UpdateOneID(agentId).SetNotes(notes).Exec(context.Background())
}

func (m *Model) GetComputerManufacturers() ([]string, error) {
	return m.Client.Computer.Query().Unique(true).Select(computer.FieldManufacturer).Strings(context.Background())
}

func (m *Model) GetComputerModels(f filters.AgentFilter) ([]string, error) {
	query := m.Client.Computer.Query().Unique(true).Select(computer.FieldModel)

	if len(f.ComputerManufacturers) > 0 {
		query.Where(computer.ManufacturerIn(f.ComputerManufacturers...))
	}

	return query.Strings(context.Background())
}

func (m *Model) CountDifferentVendor() (int, error) {
	return m.Client.Computer.Query().Select(computer.FieldManufacturer).Unique(true).Where(computer.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))).Count(context.Background())
}
