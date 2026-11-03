package models

import (
	"context"
	"testing"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileAppliesToAgent(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:agentfiles?mode=memory&_fk=1")
	defer client.Close()
	m := Model{Client: client}

	tenant1 := client.Tenant.Create().SetDescription("Tenant 1").SaveX(ctx)
	tenant2 := client.Tenant.Create().SetDescription("Tenant 2").SaveX(ctx)
	site1 := client.Site.Create().SetDescription("Site 1").SetTenant(tenant1).SaveX(ctx)
	site2 := client.Site.Create().SetDescription("Site 2").SetTenant(tenant1).SaveX(ctx)
	site3 := client.Site.Create().SetDescription("Site 3").SetTenant(tenant2).SaveX(ctx)
	tag1 := client.Tag.Create().SetTag("Accounting").SetColor("#ff0000").SetTenant(tenant1).SaveX(ctx)
	tag2 := client.Tag.Create().SetTag("Sales").SetColor("#00ff00").SetTenant(tenant1).SaveX(ctx)

	client.Agent.Create().SetID("enabled").SetHostname("pc1").SetOs("windows").SetNickname("pc1").
		SetAgentStatus(agent.AgentStatusEnabled).AddSite(site1).AddTags(tag1).ExecX(ctx)
	client.Agent.Create().SetID("waiting").SetHostname("pc2").SetOs("windows").SetNickname("pc2").
		SetAgentStatus(agent.AgentStatusWaitingForAdmission).AddSite(site1).AddTags(tag1).ExecX(ctx)

	global := client.Profile.Create().SetName("Global for all").SetApplyToAll(true).SaveX(ctx)
	tagged := client.Profile.Create().SetName("Accounting").AddTags(tag1).AddTenant(tenant1).SaveX(ctx)
	otherTag := client.Profile.Create().SetName("Sales").AddTags(tag2).AddTenant(tenant1).SaveX(ctx)
	ownSite := client.Profile.Create().SetName("Site 1").SetApplyToAll(true).AddSite(site1).SaveX(ctx)
	otherSite := client.Profile.Create().SetName("Site 2").SetApplyToAll(true).AddSite(site2).SaveX(ctx)
	otherTenant := client.Profile.Create().SetName("Tenant 2").SetApplyToAll(true).AddTenant(tenant2).AddSite(site3).SaveX(ctx)
	disabled := client.Profile.Create().SetName("Disabled").SetApplyToAll(true).SetDisabled(true).SaveX(ctx)

	for _, tc := range []struct {
		profileID int
		agentID   string
		expected  bool
	}{
		{global.ID, "enabled", true},
		{tagged.ID, "enabled", true},
		{otherTag.ID, "enabled", false},
		{ownSite.ID, "enabled", true},
		{otherSite.ID, "enabled", false},
		{otherTenant.ID, "enabled", false},
		{disabled.ID, "enabled", false},
		{global.ID, "waiting", false},
		{global.ID, "unknown", false},
		{999, "enabled", false},
	} {
		applies, err := m.ProfileAppliesToAgent(tc.profileID, tc.agentID)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, applies, "profile %d, agent %s", tc.profileID, tc.agentID)
	}
}
