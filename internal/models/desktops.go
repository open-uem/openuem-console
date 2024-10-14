package models

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/computer"
	"github.com/doncicuto/openuem_ent/deployment"
	"github.com/doncicuto/openuem_ent/operatingsystem"
)

type Desktop struct {
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

func (m *Model) CountAllDesktops() (int, error) {
	// TODO specify agent type to desktop
	count, err := m.Client.Agent.Query().Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, err
}

func mainQuery(s *sql.Selector, p partials.PaginationAndSort) {
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldHostname, agent.FieldOs, `agents."version"`, agent.FieldIP, agent.FieldMAC, operatingsystem.FieldUsername, computer.FieldManufacturer, computer.FieldModel).
		LeftJoin(sql.Table(computer.Table)).
		On(agent.FieldID, computer.OwnerColumn).
		LeftJoin(sql.Table(operatingsystem.Table)).
		On(agent.FieldID, operatingsystem.OwnerColumn).
		Limit(p.PageSize).
		Offset((p.CurrentPage - 1) * p.PageSize)
}

func (m *Model) GetDesktopsByPage(p partials.PaginationAndSort) ([]Desktop, error) {
	var err error
	var desktops []Desktop

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldHostname))
			}).Scan(context.Background(), &desktops)
		} else {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldHostname))
			}).Scan(context.Background(), &desktops)
		}
	case "os":
		if p.SortOrder == "asc" {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldOs))
			}).Scan(context.Background(), &desktops)
		} else {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldOs))
			}).Scan(context.Background(), &desktops)
		}
	case "version":
		if p.SortOrder == "asc" {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldVersion))
			}).Scan(context.Background(), &desktops)
		} else {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldVersion))
			}).Scan(context.Background(), &desktops)
		}
	case "username":
		if p.SortOrder == "asc" {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(operatingsystem.FieldUsername))
			}).Scan(context.Background(), &desktops)
		} else {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(operatingsystem.FieldUsername))
			}).Scan(context.Background(), &desktops)
		}
	case "manufacturer":
		if p.SortOrder == "asc" {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(computer.FieldManufacturer))
			}).Scan(context.Background(), &desktops)
		} else {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(computer.FieldManufacturer))
			}).Scan(context.Background(), &desktops)
		}
	case "model":
		if p.SortOrder == "asc" {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(computer.FieldModel))
			}).Scan(context.Background(), &desktops)
		} else {
			err = m.Client.Agent.Query().Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(computer.FieldModel))
			}).Scan(context.Background(), &desktops)
		}
	}

	if err != nil {
		return nil, err
	}
	return desktops, nil
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

func (m *Model) GetAntiviriInfo() ([]*ent.Agent, error) {
	agents, err := m.Client.Agent.Query().WithAntivirus().All(context.Background())
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (m *Model) GetSystemUpdatesInfo() ([]*ent.Agent, error) {
	agents, err := m.Client.Agent.Query().WithSystemupdate().All(context.Background())
	if err != nil {
		return nil, err
	}
	return agents, nil
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
