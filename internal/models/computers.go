package models

import (
	"context"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/app"
	"github.com/open-uem/ent/computer"
	"github.com/open-uem/ent/operatingsystem"
	"github.com/open-uem/ent/predicate"
	"github.com/open-uem/ent/printer"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tag"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

type Computer struct {
	ID           string
	Hostname     string `sql:"hostname"`
	Nickname     string `sql:"nickname"`
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
	SiteID       int
}

func (m *Model) CountAllComputers(f filters.AgentFilter, c *partials.CommonInfo) (int, error) {
	var query *ent.AgentQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	// Agents that haven't been admitted yet should not appear
	if siteID == -1 {
		query = m.Client.Agent.Query().
			Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = m.Client.Agent.Query().
			Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	// Apply filters
	applyComputerFilters(query, f)

	count, err := query.Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, err
}

func mainQuery(s *sql.Selector, p partials.PaginationAndSort) {
	s.Select(sql.As(agent.FieldID, "ID"), agent.FieldHostname, agent.FieldNickname, agent.FieldOs, "`t2`.`version`", agent.FieldIP, agent.FieldMAC, operatingsystem.FieldUsername, computer.FieldManufacturer, computer.FieldModel, computer.FieldSerial, agent.FieldIsRemote, agent.FieldLastContact).
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
	case "nickname":
		if p.SortOrder == "asc" {
			query = query.Order(agent.ByNickname())
		} else {
			query = query.Order(agent.ByNickname(sql.OrderDesc()))
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

func (m *Model) GetComputersByPage(p partials.PaginationAndSort, f filters.AgentFilter, c *partials.CommonInfo) ([]Computer, error) {
	var err error
	var computers []Computer
	var query *ent.AgentQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	// Agents that haven't been admitted yet should not appear
	if siteID == -1 {
		query = m.Client.Agent.Query().
			Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	} else {
		query = m.Client.Agent.Query().
			Where(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))
	}

	// Apply filters
	applyComputerFilters(query, f)

	// Apply sort
	switch p.SortBy {
	case "nickname":
		if p.SortOrder == "asc" {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Asc(agent.FieldNickname))
			}).Scan(context.Background(), &computers)
		} else {
			err = query.Modify(func(s *sql.Selector) {
				mainQuery(s, p)
				s.OrderBy(sql.Desc(agent.FieldNickname))
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
	agents, err := m.Client.Agent.Query().WithSite().WithTags().Where(agent.IDIn(sortedAgentIDs...)).All(context.Background())
	if err != nil {
		return nil, err
	}

	// Add tags and site id to each computer in order
	for i, computer := range computers {
		for _, agent := range agents {
			if computer.ID == agent.ID {
				computers[i].Tags = agent.Edges.Tags
				if len(agent.Edges.Site) == 1 {
					computers[i].SiteID = agent.Edges.Site[0].ID
				} else {
					computers[i].SiteID = -1
				}
				break
			}
		}
	}

	return computers, nil
}

func applyComputerFilters(query *ent.AgentQuery, f filters.AgentFilter) {
	if len(f.Nickname) > 0 {
		query.Where(agent.NicknameContainsFold(f.Nickname))
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

func (m *Model) GetAgentComputerInfo(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithComputer().WithMemoryslots().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	} else {
		agent, err := m.Client.Agent.Query().WithComputer().WithMemoryslots().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	}
}

func (m *Model) GetAgentOSInfo(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithOperatingsystem().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	} else {
		agent, err := m.Client.Agent.Query().WithOperatingsystem().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	}
}

func (m *Model) GetAgentNetworkAdaptersInfo(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithNetworkadapters().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	} else {
		agent, err := m.Client.Agent.Query().WithNetworkadapters().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	}
}

func (m *Model) GetAgentPrintersInfo(agentId string, c *partials.CommonInfo) ([]*ent.Printer, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		return m.Client.Printer.Query().
			Where(printer.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).
			Order(ent.Asc(printer.FieldID)).All(context.Background())
	} else {
		return m.Client.Printer.Query().
			Where(printer.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).
			Order(ent.Asc(printer.FieldID)).All(context.Background())
	}
}

func (m *Model) GetAgentLogicalDisksInfo(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithLogicaldisks().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	} else {
		agent, err := m.Client.Agent.Query().WithLogicaldisks().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	}
}

func (m *Model) GetAgentPhysicalDisksInfo(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithPhysicaldisks().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	} else {
		agent, err := m.Client.Agent.Query().WithPhysicaldisks().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	}
}

func (m *Model) GetAgentSharesInfo(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithShares().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	} else {
		agent, err := m.Client.Agent.Query().WithShares().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	}
}

func (m *Model) GetAgentMonitorsInfo(agentId string, c *partials.CommonInfo) (*ent.Agent, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		agent, err := m.Client.Agent.Query().WithMonitors().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	} else {
		agent, err := m.Client.Agent.Query().WithMonitors().WithTags().
			Where(agent.ID(agentId)).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			Only(context.Background())
		if err != nil {
			return nil, err
		}
		return agent, nil
	}
}

func (m *Model) SaveNotes(agentId string, notes string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return m.Client.Agent.UpdateOneID(agentId).
			Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).
			SetNotes(notes).Exec(context.Background())
	} else {
		return m.Client.Agent.UpdateOneID(agentId).
			Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).
			SetNotes(notes).Exec(context.Background())
	}
}

func (m *Model) GetComputerManufacturers(c *partials.CommonInfo) ([]string, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		return m.Client.Computer.Query().Where(computer.HasOwnerWith(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Unique(true).Select(computer.FieldManufacturer).Strings(context.Background())
	} else {
		return m.Client.Computer.Query().Where(computer.HasOwnerWith(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Unique(true).Select(computer.FieldManufacturer).Strings(context.Background())
	}
}

func (m *Model) GetComputerModels(f filters.AgentFilter, c *partials.CommonInfo) ([]string, error) {
	var query *ent.ComputerSelect

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Computer.Query().
			Where(computer.HasOwnerWith(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).
			Unique(true).Select(computer.FieldModel)
	} else {
		query = m.Client.Computer.Query().
			Where(computer.HasOwnerWith(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).
			Unique(true).Select(computer.FieldModel)
	}

	if len(f.ComputerManufacturers) > 0 {
		query.Where(computer.ManufacturerIn(f.ComputerManufacturers...))
	}

	return query.Strings(context.Background())
}

func (m *Model) CountDifferentVendor(c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Computer.Query().Select(computer.FieldManufacturer).Unique(true).Where(computer.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	} else {
		return m.Client.Computer.Query().Select(computer.FieldManufacturer).Unique(true).Where(computer.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	}
}

func (m *Model) SetDefaultPrinter(agentId string, printerName string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		if err := m.Client.Printer.Update().SetIsDefault(false).Where(printer.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Exec(context.Background()); err != nil {
			return err
		}
		return m.Client.Printer.Update().SetIsDefault(true).Where(printer.Name(printerName), printer.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Exec(context.Background())
	} else {
		if err := m.Client.Printer.Update().SetIsDefault(false).Where(printer.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Exec(context.Background()); err != nil {
			return err
		}
		return m.Client.Printer.Update().SetIsDefault(true).Where(printer.Name(printerName), printer.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Exec(context.Background())
	}
}

func (m *Model) RemovePrinter(agentId string, printerName string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		_, err = m.Client.Printer.Delete().
			Where(printer.Name(printerName), printer.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).
			Exec(context.Background())
		return err
	} else {
		_, err = m.Client.Printer.Delete().
			Where(printer.Name(printerName), printer.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).
			Exec(context.Background())
		return err
	}
}

func (m *Model) GetAgentAppsInfo(agentId string, c *partials.CommonInfo) ([]*ent.App, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		apps, err := m.Client.App.Query().
			Where(app.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).
			All(context.Background())
		if err != nil {
			return nil, err
		}
		return apps, nil
	} else {
		apps, err := m.Client.App.Query().
			Where(app.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).
			All(context.Background())

		if err != nil {
			return nil, err
		}
		return apps, nil
	}
}
