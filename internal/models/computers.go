package models

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/computer"
	"github.com/doncicuto/openuem_ent/deployment"
	"github.com/doncicuto/openuem_ent/metadata"
	"github.com/doncicuto/openuem_ent/operatingsystem"
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
	}

	if err != nil {
		return nil, err
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

	if len(f.OSVersions) > 0 {
		query = query.Where(agent.HasOperatingsystemWith(operatingsystem.VersionIn(f.OSVersions...)))
	}

	if len(f.ComputerManufacturers) > 0 {
		query = query.Where(agent.HasComputerWith(computer.ManufacturerIn(f.ComputerManufacturers...)))
	}

	if len(f.ComputerModels) > 0 {
		query = query.Where(agent.HasComputerWith(computer.ModelIn(f.ComputerModels...)))
	}
}

func (m *Model) GetAgentComputerInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithComputer().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentOSInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithOperatingsystem().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentNetworkAdaptersInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithNetworkadapters().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentPrintersInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithPrinters().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentLogicalDisksInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithLogicaldisks().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentSharesInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithShares().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (m *Model) GetAgentMonitorsInfo(agentId string) (*ent.Agent, error) {
	agent, err := m.Client.Agent.Query().WithMonitors().Where(agent.ID(agentId)).Only(context.Background())
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
	query := m.Client.Computer.Query().Select(computer.FieldManufacturer).Unique(true)

	data, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}

	manufacturers := []string{}

	for _, item := range data {
		manufacturers = append(manufacturers, item.Manufacturer)
	}

	return manufacturers, nil
}

func (m *Model) GetComputerModels(f filters.AgentFilter) ([]string, error) {
	query := m.Client.Computer.Query().Select(computer.FieldModel).Unique(true)

	if len(f.ComputerManufacturers) > 0 {
		query.Where(computer.ManufacturerIn(f.ComputerManufacturers...))
	}

	data, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}

	models := []string{}

	for _, item := range data {
		models = append(models, item.Model)
	}

	return models, nil
}

func (m *Model) GetOSVersions(f filters.AgentFilter) ([]string, error) {
	query := m.Client.OperatingSystem.Query().Select(operatingsystem.FieldVersion).Unique(true)

	osTypes := []string{}
	if f.WindowsAgents {
		osTypes = append(osTypes, "windows")
	}

	if f.LinuxAgents {
		osTypes = append(osTypes, "linux")
	}

	if f.MacAgents {
		osTypes = append(osTypes, "mac")
	}

	if len(osTypes) > 0 {
		query.Where(operatingsystem.TypeIn(osTypes...))
	}

	data, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}

	versions := []string{}

	for _, item := range data {
		versions = append(versions, item.Version)
	}

	return versions, nil
}
