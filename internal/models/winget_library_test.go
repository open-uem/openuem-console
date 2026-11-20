package models

import (
	"context"
	"testing"

	"github.com/open-uem/ent"
	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLibraryPackagesInCatalog(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:wingetlibrary?mode=memory&_fk=1")
	defer client.Close()
	m := Model{Client: client}

	client.SoftwarePackage.Create().SetPackageID("Example.Community").SetName("Example Tool").SetSource("winget").ExecX(ctx)
	client.SoftwarePackage.Create().SetPackageID("Example.Flatpak").SetName("Example Tool").SetSource("flatpak").ExecX(ctx)
	require.NoError(t, m.SetLibraryPackageInCatalog("ExampleCorp.Tool", "Example Tool"))

	isLibrary, err := m.IsLibraryPackage("ExampleCorp.Tool")
	require.NoError(t, err)
	assert.True(t, isLibrary)
	isLibrary, err = m.IsLibraryPackage("Example.Community")
	require.NoError(t, err)
	assert.False(t, isLibrary, "a package of winget's community repository")

	// The task form's search finds both winget sources, not other package managers
	ids := func(packages []*ent.SoftwarePackage) []string {
		out := []string{}
		for _, p := range packages {
			out = append(out, p.PackageID)
		}
		return out
	}
	found, err := m.SearchAllWingetPackages("example tool", nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Example.Community", "ExampleCorp.Tool"}, ids(found))

	// Packages of the repository that belong to other organizations are left out
	found, err = m.SearchAllWingetPackages("example tool", []string{"ExampleCorp.Tool"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Example.Community"}, ids(found))

	// The same in the Deploy search
	f := filters.DeployPackageFilter{Sources: []string{"winget", LibraryPackageSource}, ExcludePackageIDs: []string{"ExampleCorp.Tool"}}
	found, err = m.SearchPackages("example tool", partials.PaginationAndSort{SortBy: "name", SortOrder: "asc", CurrentPage: 1, PageSize: 10}, f)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Example.Community"}, ids(found))
	count, err := m.CountPackages("example tool", f)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	require.NoError(t, m.RemoveLibraryPackageFromCatalog("ExampleCorp.Tool"))
	isLibrary, err = m.IsLibraryPackage("ExampleCorp.Tool")
	require.NoError(t, err)
	assert.False(t, isLibrary, "removed from the library")
}

func TestProfileTenantIDs(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:profiletenants?mode=memory&_fk=1")
	defer client.Close()
	m := Model{Client: client}

	tenant := client.Tenant.Create().SetDescription("Org").SaveX(ctx)
	orgProfile := client.Profile.Create().SetName("org").AddTenantIDs(tenant.ID).SaveX(ctx)
	globalProfile := client.Profile.Create().SetName("global").SaveX(ctx)

	ids, err := m.ProfileTenantIDs(orgProfile.ID)
	require.NoError(t, err)
	assert.Equal(t, []int{tenant.ID}, ids)

	ids, err = m.ProfileTenantIDs(globalProfile.ID)
	require.NoError(t, err)
	assert.Empty(t, ids, "global profile")

	_, err = m.ProfileTenantIDs(globalProfile.ID + 100)
	assert.Error(t, err, "unknown profile")
}
