package models

import (
	"context"

	"github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/predicate"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/softwarepackage"
	"github.com/open-uem/ent/tenant"
)

// LibraryPackageSource is the source of the packages of the console's winget repository
// in the software package catalog used by the package search
const LibraryPackageSource = "openuem"

// EnabledWindowsAgentIDs returns the enabled Windows agents of an organization, or of
// every organization if tenantID is -1
func (m *Model) EnabledWindowsAgentIDs(tenantID int) ([]string, error) {
	query := m.Client.Agent.Query().Where(agent.AgentStatusEQ(agent.AgentStatusEnabled), agent.Os("windows"))
	if tenantID != -1 {
		query.Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))
	}
	return query.IDs(context.Background())
}

// IsLibraryPackage reports if a package id is one of the winget repository's packages
func (m *Model) IsLibraryPackage(packageID string) (bool, error) {
	return m.Client.SoftwarePackage.Query().
		Where(softwarepackage.PackageID(packageID), softwarepackage.Source(LibraryPackageSource)).
		Exist(context.Background())
}

// SetLibraryPackageInCatalog adds or renames a package of the winget repository in the
// catalog, so the package search finds it
func (m *Model) SetLibraryPackageInCatalog(packageID, name string) error {
	if err := m.RemoveLibraryPackageFromCatalog(packageID); err != nil {
		return err
	}
	return m.Client.SoftwarePackage.Create().
		SetPackageID(packageID).
		SetName(name).
		SetSource(LibraryPackageSource).
		Exec(context.Background())
}

// RemoveLibraryPackageFromCatalog removes a package of the winget repository from the catalog
func (m *Model) RemoveLibraryPackageFromCatalog(packageID string) error {
	_, err := m.Client.SoftwarePackage.Delete().
		Where(softwarepackage.PackageID(packageID), softwarepackage.Source(LibraryPackageSource)).
		Exec(context.Background())
	return err
}

// GetLibraryPackagesInCatalog returns the ids of the winget repository's packages in the catalog
func (m *Model) GetLibraryPackagesInCatalog() ([]string, error) {
	return m.Client.SoftwarePackage.Query().
		Where(softwarepackage.Source(LibraryPackageSource)).
		Select(softwarepackage.FieldPackageID).
		Strings(context.Background())
}

// ProfileTenantIDs returns the organizations of a profile, none for a global profile
func (m *Model) ProfileTenantIDs(profileID int) ([]int, error) {
	p, err := m.Client.Profile.Get(context.Background(), profileID)
	if err != nil {
		return nil, err
	}
	return p.QueryTenant().IDs(context.Background())
}

// EnabledAgentTenants returns the organizations of an enabled agent, through its sites.
// ok is false if the agent doesn't exist or isn't enabled.
func (m *Model) EnabledAgentTenants(agentID string) (tenantIDs []int, ok bool, err error) {
	a, err := m.Client.Agent.Query().
		WithSite(func(q *ent.SiteQuery) { q.WithTenant() }).
		Where(agent.ID(agentID), agent.AgentStatusEQ(agent.AgentStatusEnabled)).
		Only(context.Background())
	if ent.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	for _, s := range a.Edges.Site {
		if s.Edges.Tenant != nil {
			tenantIDs = append(tenantIDs, s.Edges.Tenant.ID)
		}
	}
	return tenantIDs, true, nil
}

// IsPackageOfOtherSource reports whether a package identifier is in the catalog with
// another source than the winget repository's, e.g. a package of winget's community
// repository. Identifiers are compared without case, like winget does.
func (m *Model) IsPackageOfOtherSource(packageID string) (bool, error) {
	return m.Client.SoftwarePackage.Query().
		Where(softwarepackage.PackageIDEqualFold(packageID), softwarepackage.SourceNEQ(LibraryPackageSource)).
		Exist(context.Background())
}

// notLibraryPackages excludes packages of the winget repository by identifier, without
// hiding packages of other sources that have the same identifier
func notLibraryPackages(packageIDs []string) predicate.SoftwarePackage {
	return softwarepackage.Not(softwarepackage.And(softwarepackage.PackageIDIn(packageIDs...), softwarepackage.Source(LibraryPackageSource)))
}
