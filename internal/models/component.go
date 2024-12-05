package models

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/component"
)

func (m *Model) CountAllUpdateServers(f filters.UpdateComponentsFilter) (int, error) {

	query := m.Client.Component.Query()

	// Apply filters
	applyComponentFilters(query, f)

	return m.Client.Component.Query().Count(context.Background())
}

func (m *Model) GetUpdateComponentsByPage(p partials.PaginationAndSort, f filters.UpdateComponentsFilter) ([]*ent.Component, error) {
	var err error
	var components []*ent.Component
	var query *ent.ComponentQuery

	query = m.Client.Component.Query().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	// Apply filters
	applyComponentFilters(query, f)

	switch p.SortBy {
	case "hostname":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(component.FieldHostname)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(component.FieldHostname)).All(context.Background())
		}
	case "component":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(component.FieldComponent)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(component.FieldComponent)).All(context.Background())
		}
	case "version":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(component.FieldVersion)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(component.FieldVersion)).All(context.Background())
		}
	case "status":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(component.FieldUpdateStatus)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(component.FieldUpdateStatus)).All(context.Background())
		}
	case "message":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(component.FieldUpdateMessage)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(component.FieldUpdateMessage)).All(context.Background())
		}
	case "when":
		if p.SortOrder == "asc" {
			components, err = query.Order(ent.Asc(component.FieldUpdateWhen)).All(context.Background())
		} else {
			components, err = query.Order(ent.Desc(component.FieldUpdateWhen)).All(context.Background())
		}
	default:
		components, err = query.Order(ent.Desc(component.FieldUpdateWhen)).All(context.Background())
	}

	if err != nil {
		return nil, err
	}
	return components, nil
}

func (m *Model) GetHigherServerReleaseInstalled() (string, error) {
	return m.Client.Component.Query().Unique(true).Order(ent.Desc(component.FieldVersion)).Select(component.FieldVersion).String(context.Background())
}

func (m *Model) SetComponent(c component.Component, version string, channel component.Channel) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	exists := true
	s, err := m.Client.Component.Query().Where(component.Hostname(hostname), component.ComponentEQ(c), component.Arch(runtime.GOARCH), component.Os(runtime.GOOS), component.Version(version), component.ChannelEQ(channel)).Only(context.Background())
	if err != nil {
		if !openuem_ent.IsNotFound(err) {
			return err
		}
		exists = false
	}

	if !exists {
		return m.Client.Component.Create().SetHostname(hostname).SetComponent(c).SetArch(runtime.GOARCH).SetOs(runtime.GOOS).SetVersion(version).SetChannel(channel).Exec(context.Background())
	}
	return m.Client.Component.Update().SetHostname(hostname).SetComponent(c).SetArch(runtime.GOARCH).SetOs(runtime.GOOS).SetVersion(version).SetChannel(channel).Where(component.ID(s.ID)).Exec(context.Background())
}

func (m *Model) GetAppliedReleases() ([]string, error) {
	return m.Client.Component.Query().Unique(true).Order(ent.Desc(component.FieldVersion)).Select(component.FieldVersion).Strings(context.Background())
}

func applyComponentFilters(query *ent.ComponentQuery, f filters.UpdateComponentsFilter) {
	if len(f.Hostname) > 0 {
		query = query.Where(component.HostnameContainsFold(f.Hostname))
	}

	if len(f.Components) > 0 {
		enumComponents := []component.Component{}
		for _, item := range f.Components {
			switch item {
			case "nats":
				enumComponents = append(enumComponents, component.ComponentNats)
			case "ocsp":
				enumComponents = append(enumComponents, component.ComponentOcsp)
			case "console":
				enumComponents = append(enumComponents, component.ComponentConsole)
			case "agent-worker":
				enumComponents = append(enumComponents, component.ComponentAgentWorker)
			case "cert-manager-worker":
				enumComponents = append(enumComponents, component.ComponentCertManagerWorker)
			case "notification-worker":
				enumComponents = append(enumComponents, component.ComponentNotificationWorker)
			case "cert-manager":
				enumComponents = append(enumComponents, component.ComponentCertManager)
			}
		}

		query = query.Where(component.ComponentIn(enumComponents...))
	}

	if len(f.Releases) > 0 {
		query = query.Where(component.VersionIn(f.Releases...))
	}

	if len(f.UpdateStatus) > 0 {
		enumStatus := []component.UpdateStatus{}
		for _, item := range f.UpdateStatus {
			switch item {
			case "Error":
				enumStatus = append(enumStatus, component.UpdateStatusError)
			case "Success":
				enumStatus = append(enumStatus, component.UpdateStatusSuccess)
			case "Pending":
				enumStatus = append(enumStatus, component.UpdateStatusPending)
			}
		}

		query = query.Where(component.UpdateStatusIn(enumStatus...))
	}

	if len(f.UpdateWhenFrom) > 0 {
		from, err := time.Parse("2006-01-02", f.UpdateWhenFrom)
		if err == nil {
			query = query.Where(component.UpdateWhenGTE(from))
		}
	}

	if len(f.UpdateWhenTo) > 0 {
		to, err := time.Parse("2006-01-02", f.UpdateWhenTo)
		if err == nil {
			query = query.Where(component.UpdateWhenLTE(to))
		}
	}
}

func (m *Model) GetAllUpdateComponents(f filters.UpdateComponentsFilter) ([]*ent.Component, error) {
	query := m.Client.Component.Query()
	// Apply filters
	applyComponentFilters(query, f)

	c, err := query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return c, nil
}
