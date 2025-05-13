package models

import (
	"context"
	"errors"
	"strconv"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/metadata"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) SaveEndpointDescription(agentID string, description string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return m.Client.Agent.Update().SetDescription(description).Where(agent.ID(agentID), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	} else {
		return m.Client.Agent.Update().SetDescription(description).Where(agent.ID(agentID), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	}
}

func (m *Model) SaveEndpointType(agentID string, endpointType string, c *partials.CommonInfo) error {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	if siteID == -1 {
		return m.Client.Agent.Update().SetEndpointType(agent.EndpointType(endpointType)).Where(agent.ID(agentID), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	} else {
		return m.Client.Agent.Update().SetEndpointType(agent.EndpointType(endpointType)).Where(agent.ID(agentID), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).Exec(context.Background())
	}
}

func (m *Model) AssociateToTenantAndSite(agentID string, newTenant, newSite string) error {
	siteID, err := strconv.Atoi(newSite)
	if err != nil {
		return err
	}
	tenantID, err := strconv.Atoi(newTenant)
	if err != nil {
		return err
	}

	// Get current agent
	a, err := m.Client.Agent.Query().WithSite().WithTags().Where(agent.ID(agentID)).Only(context.Background())
	if err != nil {
		return err
	}

	sites := a.Edges.Site
	if a.Edges.Site == nil || len(sites) != 1 {
		return errors.New("agent should be associated with one site")
	}

	currentSite := a.Edges.Site[0].ID

	s, err := m.Client.Site.Query().WithTenant().Where(site.ID(currentSite)).Only(context.Background())
	if err != nil {
		return err
	}

	if s.Edges.Tenant == nil {
		return errors.New("site should be associated with one organization")
	}

	currentTenant := s.Edges.Tenant.ID

	// if associated org changes, remove the associated metadata
	if currentTenant != tenantID {
		if _, err := m.Client.Metadata.Delete().Where(metadata.HasOwnerWith(agent.ID(agentID))).Exec(context.Background()); err != nil {
			return err
		}
	}

	query := m.Client.Agent.UpdateOneID(agentID).Where(agent.ID(agentID))

	// if associated site changes, remove the current site and add the new one
	if currentSite != siteID {
		query = query.RemoveSiteIDs(currentSite).AddSiteIDs(siteID)
	}

	// if associated org changes, remove the associated tags
	if currentTenant != tenantID {
		removeTags := []int{}
		for _, t := range a.Edges.Tags {
			removeTags = append(removeTags, t.ID)
		}

		if len(removeTags) > 0 {
			query = query.RemoveTagIDs(removeTags...)
		}
	}

	return query.Exec(context.Background())
}
