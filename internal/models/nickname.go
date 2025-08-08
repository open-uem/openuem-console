package models

import (
	"context"
	"strconv"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) SaveNickname(agentID string, nickname string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return m.Client.Agent.Update().SetNickname(nickname).Where(agent.ID(agentID), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	} else {
		return m.Client.Agent.Update().SetNickname(nickname).Where(agent.ID(agentID), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	}
}
