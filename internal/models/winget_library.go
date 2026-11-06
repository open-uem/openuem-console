package models

import (
	"context"

	"github.com/open-uem/ent/agent"
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
