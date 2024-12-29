package models

import (
	"context"

	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/operatingsystem"
)

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
