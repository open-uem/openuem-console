package models

import (
	"context"

	"github.com/open-uem/ent/agent"
)

func (m *Model) SaveEndpointDescription(agentID string, description string) error {
	return m.Client.Agent.Update().SetDescription(description).Where(agent.ID(agentID)).Exec(context.Background())
}

func (m *Model) SaveEndpointType(agentID string, endpointType string) error {
	return m.Client.Agent.Update().SetEndpointType(agent.EndpointType(endpointType)).Where(agent.ID(agentID)).Exec(context.Background())
}
