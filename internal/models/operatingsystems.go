package models

import (
	"context"

	"github.com/open-uem/openuem-console/internal/views/filters"
	ent "github.com/open-uem/openuem_ent"
	"github.com/open-uem/openuem_ent/agent"
	"github.com/open-uem/openuem_ent/operatingsystem"
)

func (m *Model) CountAgentsByOSVersion() ([]Agent, error) {
	// Info from agents waiting for admission won't be shown
	agents := []Agent{}
	err := m.Client.OperatingSystem.Query().Where(operatingsystem.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))).GroupBy(operatingsystem.FieldVersion).Aggregate(ent.Count()).Scan(context.Background(), &agents)
	if err != nil {
		return nil, err
	}
	return agents, err
}

func (m *Model) GetOSVersions(f filters.AgentFilter) ([]string, error) {
	query := m.Client.OperatingSystem.Query().Unique(true).Select(operatingsystem.FieldVersion)

	if len(f.AgentOSVersions) > 0 {
		query.Where(operatingsystem.TypeIn(f.AgentOSVersions...))
	}

	return query.Strings(context.Background())
}

func (m *Model) CountAllOSUsernames() (int, error) {
	return m.Client.OperatingSystem.Query().Select(operatingsystem.FieldUsername).Unique(true).Where(operatingsystem.HasOwnerWith(agent.AgentStatusNEQ(agent.AgentStatusWaitingForAdmission))).Count(context.Background())
}
