package models

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/app"
	"github.com/doncicuto/openuem_ent/computer"
	"github.com/doncicuto/openuem_ent/deployment"
	"github.com/doncicuto/openuem_ent/metadata"
	"github.com/doncicuto/openuem_ent/operatingsystem"
	"github.com/doncicuto/openuem_ent/predicate"
	"github.com/doncicuto/openuem_ent/tag"
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
	Tags         []*ent.Tag
}

func (m *Model) CountAllComputers(f filters.AgentFilter) (int, error) {

	query := m.Client.Agent.Query()

	// Apply filters
	applyComputerFilters(query, f)

	count, err := query.Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, err
}

func mainQuery(s *sql.Selector, p partials.PaginationAndSort) {
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldHostname, agent.FieldOs, "`t2`.`version`", agent.FieldIP, agent.FieldMAC, operatingsystem.FieldUsername, computer.FieldManufacturer, computer.FieldModel).
		LeftJoin(sql.Table(computer.Table)).
		On(agent.FieldID, computer.OwnerColumn).
		LeftJoin(sql.Table(operatingsystem.Table)).
		On(agent.FieldID, operatingsystem.OwnerColumn).
		Limit(p.PageSize).
		Offset((p.CurrentPage - 1) * p.PageSize)
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

	query := m.Client.Agent.Query()

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
		query = query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if len(f.Username) > 0 {
		query = query.Where(agent.HasOperatingsystemWith(operatingsystem.UsernameContainsFold(f.Username)))
	}

	if len(f.AgentOSVersions) > 0 {
		query = query.Where(agent.OsIn(f.AgentOSVersions...))
	}

	if len(f.OSVersions) > 0 {
		query = query.Where(agent.HasOperatingsystemWith(operatingsystem.VersionIn(f.OSVersions...)))
	}

	if len(f.ComputerManufacturers) > 0 {
		query = query.Where(agent.HasComputerWith(computer.ManufacturerIn(f.ComputerManufacturers...)))
	}

	if len(f.ComputerModels) > 0 {
		query = query.Where(agent.HasComputerWith(computer.ModelIn(f.ComputerModels...)))
	}

	if len(f.WithApplication) > 0 {
		query = query.Where(agent.HasAppsWith(app.Name(f.WithApplication)))
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

func (m *Model) GetDeploymentsForAgent(agentId string, p partials.PaginationAndSort) ([]*ent.Deployment, error) {
	query := m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.ID(agentId)))

	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(deployment.FieldName))
		} else {
			query = query.Order(ent.Desc(deployment.FieldName))
		}
	case "installation":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(deployment.FieldInstalled))
		} else {
			query = query.Order(ent.Desc(deployment.FieldInstalled))
		}
	case "updated":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(deployment.FieldUpdated))
		} else {
			query = query.Order(ent.Desc(deployment.FieldUpdated))
		}
	}

	deployments, err := query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

func (m *Model) CountDeploymentsForAgent(agentId string) (int, error) {
	return m.Client.Deployment.Query().Where(deployment.HasOwnerWith(agent.ID(agentId))).Count(context.Background())
}

func (m *Model) DeploymentAlreadyInstalled(agentId, packageId string) (bool, error) {
	return m.Client.Deployment.Query().Where(deployment.And(deployment.PackageID(packageId), deployment.HasOwnerWith(agent.ID(agentId)))).Exist(context.Background())
}

func (m *Model) GetMetadataForAgent(agentId string, p partials.PaginationAndSort) ([]*ent.Metadata, error) {
	query := m.Client.Metadata.Query().WithOrg().WithOwner().Where(metadata.HasOwnerWith(agent.ID(agentId)))

	data, err := query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *Model) CountMetadataForAgent(agentId string) (int, error) {
	return m.Client.Metadata.Query().Where(metadata.HasOwnerWith(agent.ID(agentId))).Count(context.Background())
}

func (m *Model) SaveMetadata(agentId string, metadataId int, value string) error {
	return m.Client.Metadata.Create().SetOwnerID(agentId).SetOrgID(metadataId).SetValue(value).OnConflict(sql.ConflictColumns(metadata.OwnerColumn, metadata.OrgColumn)).UpdateNewValues().Exec(context.Background())
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

func (m *Model) GetOSVersions(f filters.AgentFilter) ([]string, error) {
	query := m.Client.OperatingSystem.Query().Unique(true).Select(operatingsystem.FieldVersion)

	if len(f.AgentOSVersions) > 0 {
		query.Where(operatingsystem.TypeIn(f.AgentOSVersions...))
	}

	return query.Strings(context.Background())
}

func (m *Model) CountAllDeployments() (int, error) {
	return m.Client.Deployment.Query().Count(context.Background())
}

func (m *Model) CountAllOSUsernames() (int, error) {
	return m.Client.OperatingSystem.Query().Select(operatingsystem.FieldUsername).Unique(true).Count(context.Background())
}
