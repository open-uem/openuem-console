package models

import (
	"context"

	"github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/profile"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tag"
	"github.com/open-uem/ent/task"
	"github.com/open-uem/ent/tenant"
)

// ProfileAppliesToAgent reports if a profile is applied to an enabled agent, using the
// same rules as the worker that sends the profiles to the agents: the profile is enabled,
// it belongs to no organization and site, to the agent's organization or to one of the
// agent's sites, and it's applied to all agents or to one of the agent's tags.
func (m *Model) ProfileAppliesToAgent(profileID int, agentID string) (bool, error) {
	a, err := m.Client.Agent.Query().
		WithTags().
		WithSite(func(q *ent.SiteQuery) { q.WithTenant() }).
		Where(agent.ID(agentID), agent.AgentStatusEQ(agent.AgentStatusEnabled)).
		Only(context.Background())
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	siteIDs := []int{}
	tenantIDs := []int{}
	for _, s := range a.Edges.Site {
		siteIDs = append(siteIDs, s.ID)
		if s.Edges.Tenant != nil {
			tenantIDs = append(tenantIDs, s.Edges.Tenant.ID)
		}
	}

	tagIDs := []int{}
	for _, t := range a.Edges.Tags {
		tagIDs = append(tagIDs, t.ID)
	}

	return m.Client.Profile.Query().Where(
		profile.ID(profileID),
		profile.Disabled(false),
		profile.Or(profile.ApplyToAll(true), profile.HasTagsWith(tag.IDIn(tagIDs...))),
		profile.Or(
			profile.And(profile.Not(profile.HasTenant()), profile.Not(profile.HasSite())),
			profile.HasTenantWith(tenant.IDIn(tenantIDs...)),
			profile.HasSiteWith(site.IDIn(siteIDs...)),
		),
	).Exist(context.Background())
}

// ProfilesUsingTaskFile returns the profiles with MSI tasks whose path is the given
// uploaded file, e.g. tasks cloned into another profile
func (m *Model) ProfilesUsingTaskFile(urlPath string) ([]int, error) {
	return m.Client.Task.Query().
		Where(task.TypeIn(task.TypeMsiInstall, task.TypeMsiUninstall), task.MsiPathHasSuffix(urlPath)).
		QueryProfile().
		IDs(context.Background())
}

// GetTaskFilePaths returns the paths of the MSI tasks, used to find the uploaded files
// that are still in use
func (m *Model) GetTaskFilePaths() ([]string, error) {
	return m.Client.Task.Query().
		Where(task.TypeIn(task.TypeMsiInstall, task.TypeMsiUninstall), task.MsiPathNEQ("")).
		Select(task.FieldMsiPath).
		Strings(context.Background())
}
