package authz

import (
	"context"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/predicate"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/ent/user"
)

type AccessScope struct {
	IsAdmin   bool
	TenantIDs map[int]struct{}
	SiteIDs   map[int]struct{}
	AgentIDs  map[string]struct{}
}

func (s *AccessScope) AllowsTenant(tenantID int) bool {
	if s == nil {
		return false
	}
	if s.IsAdmin {
		return true
	}
	_, ok := s.TenantIDs[tenantID]
	return ok
}

func (s *AccessScope) AllowsSite(siteID int) bool {
	if s == nil {
		return false
	}
	if s.IsAdmin {
		return true
	}
	_, ok := s.SiteIDs[siteID]
	return ok
}

func (s *AccessScope) AllowsAgent(agentID string) bool {
	if s == nil {
		return false
	}
	if s.IsAdmin {
		return true
	}
	_, ok := s.AgentIDs[agentID]
	return ok
}

func LoadScope(ctx context.Context, client *ent.Client, userID string) (*AccessScope, error) {
	u, err := client.User.Query().
		Where(user.ID(userID)).
		WithAllowedTenants().
		WithAllowedSites(func(q *ent.SiteQuery) { q.WithTenant() }).
		WithAllowedAgents(func(q *ent.AgentQuery) { q.WithSite(func(sq *ent.SiteQuery) { sq.WithTenant() }) }).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	scope := &AccessScope{
		IsAdmin:   u.ConsoleRole == user.ConsoleRoleAdmin,
		TenantIDs: make(map[int]struct{}),
		SiteIDs:   make(map[int]struct{}),
		AgentIDs:  make(map[string]struct{}),
	}

	for _, t := range u.Edges.AllowedTenants {
		scope.TenantIDs[t.ID] = struct{}{}
	}

	for _, s := range u.Edges.AllowedSites {
		scope.SiteIDs[s.ID] = struct{}{}
		if s.Edges.Tenant != nil {
			scope.TenantIDs[s.Edges.Tenant.ID] = struct{}{}
		}
	}

	for _, a := range u.Edges.AllowedAgents {
		scope.AgentIDs[a.ID] = struct{}{}
		for _, s := range a.Edges.Site {
			scope.SiteIDs[s.ID] = struct{}{}
			if s.Edges.Tenant != nil {
				scope.TenantIDs[s.Edges.Tenant.ID] = struct{}{}
			}
		}
	}

	return scope, nil
}

func AgentScopePredicates(scope *AccessScope) []predicate.Agent {
	if scope == nil || scope.IsAdmin {
		return nil
	}

	preds := make([]predicate.Agent, 0, 3)
	if len(scope.TenantIDs) > 0 {
		tenantIDs := make([]int, 0, len(scope.TenantIDs))
		for id := range scope.TenantIDs {
			tenantIDs = append(tenantIDs, id)
		}
		preds = append(preds, agent.HasSiteWith(site.HasTenantWith(tenant.IDIn(tenantIDs...))))
	}
	if len(scope.SiteIDs) > 0 {
		siteIDs := make([]int, 0, len(scope.SiteIDs))
		for id := range scope.SiteIDs {
			siteIDs = append(siteIDs, id)
		}
		preds = append(preds, agent.HasSiteWith(site.IDIn(siteIDs...)))
	}
	if len(scope.AgentIDs) > 0 {
		agentIDs := make([]string, 0, len(scope.AgentIDs))
		for id := range scope.AgentIDs {
			agentIDs = append(agentIDs, id)
		}
		preds = append(preds, agent.IDIn(agentIDs...))
	}
	return preds
}
