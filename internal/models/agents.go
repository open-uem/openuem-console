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
	"github.com/doncicuto/openuem_ent/antivirus"
	"github.com/doncicuto/openuem_ent/computer"
	"github.com/doncicuto/openuem_ent/operatingsystem"
	"github.com/doncicuto/openuem_ent/predicate"
	"github.com/doncicuto/openuem_ent/printer"
	"github.com/doncicuto/openuem_ent/release"
	"github.com/doncicuto/openuem_ent/systemupdate"
	"github.com/doncicuto/openuem_ent/tag"
	"github.com/doncicuto/openuem_nats"
)

type Agent struct {
	OS      string
	Version string
	Status  string
	Count   int
}

func (m *Model) GetAllAgents(f filters.AgentFilter) ([]*ent.Agent, error) {
	// Info from agents waiting for admission won't be shown

	query := m.Client.Agent.Query().Where(agent.StatusNEQ(agent.StatusWaitingForAdmission))
	// Apply filters
	applyAgentFilters(query, f)

	agents, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (m *Model) GetAgentsByPage(p partials.PaginationAndSort, f filters.AgentFilter, excludeWaitingForAdmissionAgents bool) ([]*ent.Agent, error) {
	var err error
	var apps []*ent.Agent
	var query *ent.AgentQuery

	// Info from agents waiting for admission won't be shown
	if excludeWaitingForAdmissionAgents {
		query = m.Client.Agent.Query().Where(agent.StatusNEQ(agent.StatusWaitingForAdmission)).WithTags().WithRelease().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
	} else {
		query = m.Client.Agent.Query().WithTags().WithRelease().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)
	}

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
			apps, err = query.Order(agent.ByReleaseField(release.FieldVersion, sql.OrderAsc())).All(context.Background())
		} else {
			apps, err = query.Order(agent.ByReleaseField(release.FieldVersion, sql.OrderDesc())).All(context.Background())
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
	case "status":
		if p.SortOrder == "asc" {
			apps, err = query.Order(ent.Asc(agent.FieldStatus)).All(context.Background())
		} else {
			apps, err = query.Order(ent.Desc(agent.FieldStatus)).All(context.Background())
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
	agent, err := m.Client.Agent.Query().WithTags().WithComputer().WithOperatingsystem().Where(agent.ID(agentId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return agent, err
}

func (m *Model) CountAgentsByOS() ([]Agent, error) {
	// Info from agents waiting for admission won't be shown
	agents := []Agent{}
	err := m.Client.Agent.Query().Modify(func(s *sql.Selector) {
		s.Select(agent.FieldOs, sql.As(sql.Count("os"), "count")).Where(sql.And(sql.NEQ(agent.FieldStatus, agent.StatusWaitingForAdmission))).GroupBy("os").OrderBy("count")
	}).Scan(context.Background(), &agents)
	if err != nil {
		return nil, err
	}
	return agents, err
}

func (m *Model) CountAllAgents(f filters.AgentFilter, excludeWaitingForAdmissionAgents bool) (int, error) {
	var query *ent.AgentQuery

	// Info from agents waiting for admission won't be shown
	if excludeWaitingForAdmissionAgents {
		query = m.Client.Agent.Query().Where(agent.StatusNEQ(agent.StatusWaitingForAdmission))
	} else {
		query = m.Client.Agent.Query()
	}

	applyAgentFilters(query, f)

	count, err := query.Count(context.Background())
	return count, err
}

func (m *Model) GetAgentsUsedOSes() ([]string, error) {
	return m.Client.Agent.Query().Unique(true).Select(agent.FieldOs).Strings(context.Background())
}

func applyAgentFilters(query *ent.AgentQuery, f filters.AgentFilter) {
	if len(f.Hostname) > 0 {
		query = query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if len(f.AgentStatusOptions) > 0 {
		if len(f.AgentStatusOptions) == 1 && f.AgentStatusOptions[0] == "WaitingForAdmission" {
			query = query.Where(agent.StatusEQ(agent.StatusWaitingForAdmission))
		}

		if len(f.AgentStatusOptions) == 1 && f.AgentStatusOptions[0] == "Enabled" {
			query = query.Where(agent.StatusEQ(agent.StatusEnabled))
		}

		if len(f.AgentStatusOptions) == 1 && f.AgentStatusOptions[0] == "Disabled" {
			query = query.Where(agent.StatusEQ(agent.StatusDisabled))
		}
	}

	/* if len(f.Versions) > 0 {
		query = query.Where(agent.VersionIn(f.Versions...))
	} */

	if len(f.AgentOSVersions) > 0 {
		query = query.Where(agent.OsIn(f.AgentOSVersions...))
	}

	if len(f.ContactFrom) > 0 {
		from, err := time.Parse("2006-01-02", f.ContactFrom)
		if err == nil {
			query = query.Where(agent.LastContactGTE(from))
		}
	}

	if len(f.ContactTo) > 0 {
		to, err := time.Parse("2006-01-02", f.ContactTo)
		if err == nil {
			query = query.Where(agent.LastContactLTE(to))
		}
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
	// Info from agents waiting for admission won't be shown
	agents := []Agent{}
	err := m.Client.OperatingSystem.Query().Where(operatingsystem.HasOwnerWith(agent.StatusNEQ(agent.StatusWaitingForAdmission))).GroupBy(operatingsystem.FieldVersion).Aggregate(ent.Count()).Scan(context.Background(), &agents)
	if err != nil {
		return nil, err
	}
	return agents, err
}

func (m *Model) CountAgentsReportedLast24h() (int, error) {
	count, err := m.Client.Agent.Query().Where(agent.LastContactGTE(time.Now().AddDate(0, 0, -1)), agent.StatusNEQ(agent.StatusWaitingForAdmission)).Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, err
}

func (m *Model) CountAgentsNotReportedLast24h() (int, error) {
	count, err := m.Client.Agent.Query().Where(agent.LastContactLT(time.Now().AddDate(0, 0, -1)), agent.StatusNEQ(agent.StatusWaitingForAdmission)).Count(context.Background())
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
	_, err := m.Client.Agent.UpdateOneID(agentId).SetStatus(agent.StatusEnabled).Save(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) DisableAgent(agentId string) error {
	_, err := m.Client.Agent.UpdateOneID(agentId).SetStatus(agent.StatusDisabled).Save(context.Background())
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

func (m *Model) CountPendingUpdateAgents() (int, error) {
	return m.Client.Agent.Query().Where(agent.HasSystemupdateWith(systemupdate.PendingUpdatesEQ(true)), agent.StatusNEQ(agent.StatusWaitingForAdmission)).Count(context.Background())
}

func (m *Model) CountDisabledAntivirusAgents() (int, error) {
	return m.Client.Agent.Query().Where(agent.HasAntivirusWith(antivirus.IsActive(false)), agent.StatusNEQ(agent.StatusWaitingForAdmission)).Count(context.Background())
}

func (m *Model) CountOutdatedAntivirusDatabaseAgents() (int, error) {
	return m.Client.Agent.Query().Where(agent.HasAntivirusWith(antivirus.IsUpdated(false)), agent.StatusNEQ(agent.StatusWaitingForAdmission)).Count(context.Background())
}

func (m *Model) CountNoAutoupdateAgents() (int, error) {
	return m.Client.Agent.Query().Where(agent.HasSystemupdateWith(systemupdate.Not(systemupdate.StatusContains(openuem_nats.NOTIFY_SCHEDULED_INSTALLATION))), agent.StatusNEQ(agent.StatusWaitingForAdmission)).Count(context.Background())
}

func (m *Model) CountVNCSupportedAgents() (int, error) {
	return m.Client.Agent.Query().Where(agent.Not(agent.Vnc("")), agent.StatusNEQ(agent.StatusWaitingForAdmission)).Count(context.Background())
}

func (m *Model) CountDifferentVendor() (int, error) {
	return m.Client.Computer.Query().Select(computer.FieldManufacturer).Unique(true).Where(computer.HasOwnerWith(agent.StatusNEQ(agent.StatusWaitingForAdmission))).Count(context.Background())
}

func (m *Model) CountDifferentPrinters() (int, error) {
	return m.Client.Printer.Query().Select(printer.FieldName).Unique(true).Where(printer.HasOwnerWith(agent.StatusNEQ(agent.StatusWaitingForAdmission))).Count(context.Background())
}

func (m *Model) CountDisabledAgents() (int, error) {
	return m.Client.Agent.Query().Where(agent.StatusEQ(agent.StatusDisabled)).Count(context.Background())
}

func (m *Model) CountWaitingForAdmissionAgents() (int, error) {
	return m.Client.Agent.Query().Where(agent.StatusEQ(agent.StatusWaitingForAdmission)).Count(context.Background())
}

func (m *Model) SaveAgentUpdateInfo(agentId, status, description, version string) error {
	return m.Client.Agent.UpdateOneID(agentId).
		SetUpdateTaskStatus(status).
		SetUpdateTaskDescription(description).
		SetUpdateTaskExecution(time.Time{}).
		SetUpdateTaskVersion(version).
		SetUpdateTaskResult("").Exec(context.Background())
}

func (m *Model) GetUpdateAgentsByPage(p partials.PaginationAndSort, f filters.UpdateAgentsFilter) ([]*ent.Agent, error) {
	var err error
	var agents []*ent.Agent

	query := m.Client.Agent.Query().Where(agent.StatusNEQ(agent.StatusWaitingForAdmission)).WithTags().WithRelease().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	// Apply filters
	applyUpdateAgentsFilters(query, f)

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldHostname)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldHostname)).All(context.Background())
		}
	case "version":
		if p.SortOrder == "asc" {
			agents, err = query.Order(agent.ByReleaseField(release.FieldVersion, sql.OrderAsc())).All(context.Background())
		} else {
			agents, err = query.Order(agent.ByReleaseField(release.FieldVersion, sql.OrderDesc())).All(context.Background())
		}
	case "taskStatus":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldUpdateTaskStatus)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskStatus)).All(context.Background())
		}
	case "taskDescription":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldUpdateTaskDescription)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskDescription)).All(context.Background())
		}
	case "taskLastExecution":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldUpdateTaskExecution)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskExecution)).All(context.Background())
		}
	case "taskResult":
		if p.SortOrder == "asc" {
			agents, err = query.Order(ent.Asc(agent.FieldUpdateTaskResult)).All(context.Background())
		} else {
			agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskResult)).All(context.Background())
		}
	default:
		agents, err = query.Order(ent.Desc(agent.FieldUpdateTaskExecution)).All(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (m *Model) CountAllUpdateAgents(f filters.UpdateAgentsFilter) (int, error) {
	query := m.Client.Agent.Query().Where(agent.StatusNEQ(agent.StatusWaitingForAdmission))

	applyUpdateAgentsFilters(query, f)

	count, err := query.Count(context.Background())
	return count, err
}

func (m *Model) GetAllUpdateAgents(f filters.UpdateAgentsFilter) ([]*ent.Agent, error) {
	query := m.Client.Agent.Query().Where(agent.StatusNEQ(agent.StatusWaitingForAdmission))
	// Apply filters
	applyUpdateAgentsFilters(query, f)

	agents, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func applyUpdateAgentsFilters(query *ent.AgentQuery, f filters.UpdateAgentsFilter) {
	if len(f.Hostname) > 0 {
		query = query.Where(agent.HostnameContainsFold(f.Hostname))
	}

	if len(f.Releases) > 0 {
		query = query.Where(agent.HasReleaseWith(release.VersionIn(f.Releases...)))
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

	if len(f.TaskStatus) > 0 {
		query = query.Where(agent.UpdateTaskStatusIn(f.TaskStatus...))
	}

	if len(f.TaskResult) > 0 {
		query = query.Where(agent.UpdateTaskResultContainsFold(f.TaskResult))
	}

	if len(f.TaskLastExecutionFrom) > 0 {
		from, err := time.Parse("2006-01-02", f.TaskLastExecutionFrom)
		if err == nil {
			query = query.Where(agent.UpdateTaskExecutionGTE(from))
		}
	}

	if len(f.TaskLastExecutionTo) > 0 {
		to, err := time.Parse("2006-01-02", f.TaskLastExecutionTo)
		if err == nil {
			query = query.Where(agent.UpdateTaskExecutionLTE(to))
		}
	}
}
