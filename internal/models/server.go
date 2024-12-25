package models

import (
	"context"
	"time"

	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/release"
	"github.com/doncicuto/openuem_ent/server"
)

func (m *Model) GetLatestServerRelease(channel string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.Channel(channel), release.ReleaseTypeEQ(release.ReleaseTypeServer)).Order(ent.Desc(release.FieldVersion)).First(context.Background())
}

func (m *Model) GetServerReleases() ([]string, error) {
	return m.Client.Release.Query().Unique(true).Order(ent.Desc(release.FieldVersion)).Where(release.ReleaseTypeEQ(release.ReleaseTypeServer)).Select(release.FieldVersion).Strings(context.Background())
}

func (m *Model) CountAllUpdateServers(f filters.UpdateServersFilter) (int, error) {

	query := m.Client.Server.Query()

	// Apply filters
	applyServerFilters(query, f)

	return m.Client.Server.Query().Count(context.Background())
}

func (m *Model) GetUpdateServersByPage(p partials.PaginationAndSort, f filters.UpdateServersFilter) ([]*ent.Server, error) {
	var err error
	var components []*ent.Server
	var query *ent.ServerQuery

	query = m.Client.Server.Query().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	// Apply filters
	applyServerFilters(query, f)

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(server.FieldHostname)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(server.FieldHostname)).All(context.Background())
		}
	case "version":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(server.FieldVersion)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(server.FieldVersion)).All(context.Background())
		}
	case "status":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(server.FieldUpdateStatus)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(server.FieldUpdateStatus)).All(context.Background())
		}
	case "message":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(server.FieldUpdateMessage)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(server.FieldUpdateMessage)).All(context.Background())
		}
	case "when":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(server.FieldUpdateWhen)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(server.FieldUpdateWhen)).All(context.Background())
		}
	default:
		components, err = query.Order(ent.Desc(server.FieldUpdateWhen)).All(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return components, nil
}

func (m *Model) GetHigherServerReleaseInstalled() (*ent.Server, error) {
	return m.Client.Server.Query().Unique(true).Order(ent.Desc(server.FieldVersion)).Select(server.FieldVersion).First(context.Background())
}

func (m *Model) GetAppliedReleases() ([]string, error) {
	return m.Client.Server.Query().Unique(true).Order(ent.Desc(server.FieldVersion)).Select(server.FieldVersion).Strings(context.Background())
}

func applyServerFilters(query *ent.ServerQuery, f filters.UpdateServersFilter) {
	if len(f.Hostname) > 0 {
		query = query.Where(server.HostnameContainsFold(f.Hostname))
	}

	if len(f.Releases) > 0 {
		query = query.Where(server.VersionIn(f.Releases...))
	}

	if len(f.UpdateStatus) > 0 {
		enumStatus := []server.UpdateStatus{}
		for _, item := range f.UpdateStatus {
			switch item {
			case "Error":
				enumStatus = append(enumStatus, server.UpdateStatusError)
			case "Success":
				enumStatus = append(enumStatus, server.UpdateStatusSuccess)
			case "Pending":
				enumStatus = append(enumStatus, server.UpdateStatusPending)
			}
		}

		query = query.Where(server.UpdateStatusIn(enumStatus...))
	}

	if len(f.UpdateWhenFrom) > 0 {
		from, err := time.Parse("2006-01-02", f.UpdateWhenFrom)
		if err == nil {
			query = query.Where(server.UpdateWhenGTE(from))
		}
	}

	if len(f.UpdateWhenTo) > 0 {
		to, err := time.Parse("2006-01-02", f.UpdateWhenTo)
		if err == nil {
			query = query.Where(server.UpdateWhenLTE(to))
		}
	}
}

func (m *Model) GetAllUpdateServers(f filters.UpdateServersFilter) ([]*ent.Server, error) {
	query := m.Client.Server.Query()
	// Apply filters
	applyServerFilters(query, f)

	c, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (m *Model) SaveServerUpdateInfo(serverId int, status server.UpdateStatus, description, version string) error {
	return m.Client.Server.UpdateOneID(serverId).
		SetUpdateStatus(status).
		SetUpdateMessage(description).
		SetUpdateWhen(time.Time{}).
		SetVersion(version).
		Exec(context.Background())
}

func (m *Model) GetServerById(serverId int) (*ent.Server, error) {
	server, err := m.Client.Server.Query().Where(server.ID(serverId)).Only(context.Background())
	if err != nil {
		return nil, err
	}
	return server, err
}

func (m *Model) GetServersReleaseByType(release_type release.ReleaseType, channel, os, arch, version string) (*openuem_ent.Release, error) {
	return m.Client.Release.Query().Where(release.ReleaseTypeEQ(release_type), release.Channel(channel), release.Os(os), release.Arch(arch), release.Version(version)).Only(context.Background())
}

func (m *Model) DeleteServer(serverId int) error {
	return m.Client.Server.DeleteOneID(serverId).Exec(context.Background())
}

func (m *Model) ServersExists() (bool, error) {
	return m.Client.Server.Query().Exist(context.Background())
}
