package models

import (
	"context"
	"errors"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/ent/user"
	"github.com/open-uem/openuem-console/internal/authz"
)

func (m *Model) GetUserWithPermissions(uid string) (*ent.User, error) {
	return m.Client.User.Query().
		Where(user.ID(uid)).
		WithAllowedTenants().
		WithAllowedSites().
		WithAllowedAgents().
		Only(context.Background())
}

func (m *Model) GetAllSitesForPermissions() ([]*ent.Site, error) {
	return m.Client.Site.Query().WithTenant().All(context.Background())
}

func (m *Model) GetAllAgentsForPermissions(limit int) ([]*ent.Agent, error) {
	query := m.Client.Agent.Query().WithSite(func(sq *ent.SiteQuery) { sq.WithTenant() })
	if limit > 0 {
		query = query.Limit(limit)
	}
	return query.All(context.Background())
}

func (m *Model) GetTenantsForScope(scope *authz.AccessScope) ([]*ent.Tenant, error) {
	query := m.Client.Tenant.Query()
	if scope != nil && !scope.IsAdmin {
		ids := make([]int, 0, len(scope.TenantIDs))
		for id := range scope.TenantIDs {
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			return []*ent.Tenant{}, nil
		}
		query = query.Where(tenant.IDIn(ids...))
	}
	return query.All(context.Background())
}

func (m *Model) GetTenantByIDForScope(tenantID int, scope *authz.AccessScope) (*ent.Tenant, error) {
	if scope != nil && !scope.IsAdmin && !scope.AllowsTenant(tenantID) {
		return nil, errors.New("tenant access denied")
	}
	return m.Client.Tenant.Query().Where(tenant.ID(tenantID)).Only(context.Background())
}

func (m *Model) GetDefaultTenantForScope(scope *authz.AccessScope) (*ent.Tenant, error) {
	if scope != nil && !scope.IsAdmin {
		tenants, err := m.GetTenantsForScope(scope)
		if err != nil {
			return nil, err
		}
		if len(tenants) == 0 {
			return nil, errors.New("no accessible tenant")
		}
		return tenants[0], nil
	}
	return m.GetDefaultTenant()
}

func (m *Model) GetAssociatedSitesForScope(t *ent.Tenant, scope *authz.AccessScope) ([]*ent.Site, error) {
	query := m.Client.Site.Query().Where(site.HasTenantWith(tenant.ID(t.ID)))
	if scope != nil && !scope.IsAdmin {
		if !scope.AllowsTenant(t.ID) {
			ids := make([]int, 0, len(scope.SiteIDs))
			for id := range scope.SiteIDs {
				ids = append(ids, id)
			}
			if len(ids) == 0 {
				return []*ent.Site{}, nil
			}
			query = query.Where(site.IDIn(ids...))
		}
	}
	return query.All(context.Background())
}

func (m *Model) GetDefaultSiteForScope(t *ent.Tenant, scope *authz.AccessScope) (*ent.Site, error) {
	sites, err := m.GetAssociatedSitesForScope(t, scope)
	if err != nil {
		return nil, err
	}
	for _, s := range sites {
		if s.IsDefault {
			return s, nil
		}
	}
	if len(sites) > 0 {
		return sites[0], nil
	}
	return nil, errors.New("no accessible site")
}

func (m *Model) GetSiteByIdForScope(tenantID int, siteID int, scope *authz.AccessScope) (*ent.Site, error) {
	if scope != nil && !scope.IsAdmin && !(scope.AllowsTenant(tenantID) || scope.AllowsSite(siteID)) {
		return nil, errors.New("site access denied")
	}
	return m.GetSiteById(tenantID, siteID)
}

func (m *Model) ApplyAgentScope(query *ent.AgentQuery, scope *authz.AccessScope) *ent.AgentQuery {
	if scope == nil || scope.IsAdmin {
		return query
	}
	preds := authz.AgentScopePredicates(scope)
	if len(preds) == 0 {
		return query.Where(agent.ID("__no_access__"))
	}
	return query.Where(agent.Or(preds...))
}

func (m *Model) SaveUserPermissions(uid string, role user.ConsoleRole, tenantIDs []int, siteIDs []int, agentIDs []string) error {
	// clear and set in one update
	u := m.Client.User.UpdateOneID(uid).SetConsoleRole(role).ClearAllowedTenants().ClearAllowedSites().ClearAllowedAgents()
	if len(tenantIDs) > 0 {
		u = u.AddAllowedTenantIDs(tenantIDs...)
	}
	if len(siteIDs) > 0 {
		u = u.AddAllowedSiteIDs(siteIDs...)
	}
	if len(agentIDs) > 0 {
		u = u.AddAllowedAgentIDs(agentIDs...)
	}
	return u.Exec(context.Background())
}
