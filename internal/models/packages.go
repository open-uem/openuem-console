package models

import (
	"context"

	_ "github.com/mattn/go-sqlite3"
	"github.com/open-uem/ent"
	"github.com/open-uem/ent/softwarepackage"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) SearchPackages(packageName string, p partials.PaginationAndSort, f filters.DeployPackageFilter) ([]*ent.SoftwarePackage, error) {
	sources := []string{}
	for _, s := range f.Sources {
		sources = append(sources, s)
	}

	query := m.Client.SoftwarePackage.Query().Where(softwarepackage.NameContainsFold(packageName))

	// Query the SQLite database
	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			if len(f.Sources) == 0 {
				query.Order(ent.Asc(softwarepackage.FieldName))
			} else {
				query.Where(softwarepackage.SourceIn(sources...)).Order(ent.Asc(softwarepackage.FieldName))
			}
		} else {
			if len(f.Sources) == 0 {
				query.Order(ent.Desc(softwarepackage.FieldName))
			} else {
				query.Where(softwarepackage.SourceIn(sources...)).Order(ent.Desc(softwarepackage.FieldName))
			}
		}
	case "source":
		if p.SortOrder == "asc" {
			if len(f.Sources) == 0 {
				query.Order(ent.Asc(softwarepackage.FieldSource))
			} else {
				query.Where(softwarepackage.SourceIn(sources...)).Order(ent.Asc(softwarepackage.FieldSource))
			}
		} else {
			if len(f.Sources) == 0 {
				query.Order(ent.Desc(softwarepackage.FieldSource))
			} else {
				query.Where(softwarepackage.SourceIn(sources...)).Order(ent.Desc(softwarepackage.FieldSource))
			}
		}
	}

	return query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
}

func (m *Model) CountPackages(packageName string, f filters.DeployPackageFilter) (int, error) {
	sources := []string{}
	for _, s := range f.Sources {
		sources = append(sources, s)
	}

	query := m.Client.SoftwarePackage.Query().Where(softwarepackage.NameContainsFold(packageName))

	if len(f.Sources) != 0 {
		query.Where(softwarepackage.SourceIn(sources...))
	}

	return query.Count(context.Background())
}

func (m *Model) SearchAllWingetPackages(packageName string) ([]*ent.SoftwarePackage, error) {
	return m.Client.SoftwarePackage.Query().Where(softwarepackage.NameContainsFold(packageName), softwarepackage.Source("winget")).All(context.Background())
}

func (m *Model) SearchAllFlatpakPackages(packageName string) ([]*ent.SoftwarePackage, error) {
	return m.Client.SoftwarePackage.Query().Where(softwarepackage.NameContainsFold(packageName), softwarepackage.Source("flatpak")).All(context.Background())
}

func (m *Model) SearchAllHomeBrewFormulaePackages(packageName string) ([]*ent.SoftwarePackage, error) {
	return m.Client.SoftwarePackage.Query().Where(softwarepackage.NameContainsFold(packageName), softwarepackage.Source("brew"), softwarepackage.BrewType("formula")).All(context.Background())
}

func (m *Model) SearchAllHomeBrewCasksPackages(packageName string) ([]*ent.SoftwarePackage, error) {
	return m.Client.SoftwarePackage.Query().Where(softwarepackage.NameContainsFold(packageName), softwarepackage.Source("brew"), softwarepackage.BrewType("cask")).All(context.Background())
}
