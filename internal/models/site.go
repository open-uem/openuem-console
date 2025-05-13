package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) CreateDefaultSite(tenant *ent.Tenant) (*ent.Site, error) {
	return m.Client.Site.Create().SetDescription("DefaultSite").SetIsDefault(true).SetTenantID(tenant.ID).Save(context.Background())
}

func (m *Model) CountSites(tenantID int) (int, error) {
	return m.Client.Site.Query().Where(site.HasTenantWith(tenant.ID(tenantID))).Count(context.Background())
}

func (m *Model) GetDefaultSite(t *ent.Tenant) (*ent.Site, error) {
	return m.Client.Site.Query().Where(site.IsDefault(true), site.HasTenantWith(tenant.ID(t.ID))).Only(context.Background())
}

func (m *Model) GetAssociatedSites(t *ent.Tenant) ([]*ent.Site, error) {
	return m.Client.Site.Query().Where(site.HasTenantWith(tenant.ID(t.ID))).All(context.Background())
}

func (m *Model) GetSite(siteID int) (*ent.Site, error) {
	return m.Client.Site.Query().WithTenant().Where(site.ID(siteID)).Only(context.Background())
}

func (m *Model) GetSiteById(tenantID int, siteID int) (*ent.Site, error) {
	return m.Client.Site.Query().Where(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))).Only(context.Background())
}

func (m *Model) GetSites(tenantID int) ([]*ent.Site, error) {
	return m.Client.Site.Query().Where(site.HasTenantWith(tenant.ID(tenantID))).All(context.Background())
}

func (m *Model) GetSitesByPage(p partials.PaginationAndSort, f filters.SiteFilter, tenantID string) ([]*ent.Site, error) {
	id, err := strconv.Atoi(tenantID)
	if err != nil {
		return nil, err
	}

	query := m.Client.Site.Query().Where(site.HasTenantWith(tenant.ID(id)))

	applySitesFilter(query, f)

	switch p.SortBy {
	case "ID":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(site.FieldID))
		} else {
			query.Order(ent.Desc(site.FieldID))
		}
	case "name":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(site.FieldDescription))
		} else {
			query.Order(ent.Desc(site.FieldDescription))
		}
	case "default":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(site.FieldIsDefault))
		} else {
			query.Order(ent.Desc(site.FieldIsDefault))
		}
	case "created":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(site.FieldCreated))
		} else {
			query.Order(ent.Desc(site.FieldCreated))
		}
	case "modified":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(site.FieldModified))
		} else {
			query.Order(ent.Desc(site.FieldModified))
		}

	default:
		query.Order(ent.Asc(site.FieldID))
	}

	return query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
}

func (m *Model) CountAllSites(f filters.SiteFilter, tenantID string) (int, error) {
	id, err := strconv.Atoi(tenantID)
	if err != nil {
		return 0, err
	}

	query := m.Client.Site.Query().Where(site.HasTenantWith(tenant.ID(id)))

	applySitesFilter(query, f)

	count, err := query.Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, nil
}

func applySitesFilter(query *ent.SiteQuery, f filters.SiteFilter) {
	if len(f.Name) > 0 {
		query.Where(site.DescriptionContainsFold(f.Name))
	}

	if len(f.Domain) > 0 {
		query.Where(site.DomainContainsFold(f.Domain))
	}

	if len(f.CreatedFrom) > 0 {
		dateFrom, err := time.Parse("2006-01-02", f.CreatedFrom)
		if err == nil {
			query.Where(site.CreatedGTE(dateFrom))
		}
	}

	if len(f.CreatedTo) > 0 {
		dateTo, err := time.Parse("2006-01-02", f.CreatedTo)
		if err == nil {
			query.Where(site.CreatedLTE(dateTo))
		}
	}

	if len(f.ModifiedFrom) > 0 {
		dateFrom, err := time.Parse("2006-01-02", f.ModifiedFrom)
		if err == nil {
			query.Where(site.ModifiedGTE(dateFrom))
		}
	}

	if len(f.ModifiedTo) > 0 {
		dateTo, err := time.Parse("2006-01-02", f.ModifiedTo)
		if err == nil {
			query.Where(site.ModifiedLTE(dateTo))
		}
	}

	if len(f.DefaultOptions) > 0 {
		if len(f.DefaultOptions) == 1 && f.DefaultOptions[0] == "Yes" {
			query.Where(site.IsDefaultEQ(true))
		}

		if len(f.DefaultOptions) == 1 && f.DefaultOptions[0] == "No" {
			query.Where(site.IsDefaultEQ(false))
		}
	}
}

func (m *Model) AddSite(tenantID int, name string, isDefault bool, domain string) error {
	if isDefault {
		// Remove the is default property for existing sites
		if err := m.Client.Site.Update().Where(site.HasTenantWith(tenant.ID(tenantID))).SetIsDefault(false).Exec(context.Background()); err != nil {
			return err
		}
	}

	return m.Client.Site.Create().SetDescription(name).SetIsDefault(isDefault).SetDomain(domain).SetTenantID(tenantID).Exec(context.Background())
}

func (m *Model) UpdateSite(tenantID int, siteID int, desc string, domain string, isDefault bool) error {

	query := m.Client.Site.Update().Where(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))).SetDescription(desc).SetDomain(domain)

	if isDefault {
		if err := m.Client.Site.Update().Where(site.Not(site.ID(siteID)), site.HasTenantWith(tenant.ID(tenantID))).SetIsDefault(false).Exec(context.Background()); err != nil {
			return err
		}
		return query.SetIsDefault(true).Exec(context.Background())
	} else {
		count, err := m.Client.Site.Query().Where(site.Not(site.ID(siteID)), site.HasTenantWith(tenant.ID(tenantID)), site.IsDefault(true)).Count(context.Background())
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("this is the current default site, you cannot remove it as default site until you select a new default site first")
		}
		return query.SetIsDefault(false).Exec(context.Background())
	}
}

func (m *Model) DeleteSite(tenantID int, siteID int) error {
	_, err := m.Client.Site.Delete().Where(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))).Exec(context.Background())
	return err
}

func (m *Model) SiteNameTaken(tenantID int, desc string) (bool, error) {
	return m.Client.Site.Query().Where(site.HasTenantWith(tenant.ID(tenantID)), site.Description(desc)).Exist(context.Background())
}

func (m *Model) GetAgentsBySite(tenantID int, siteID int) ([]*ent.Agent, error) {
	return m.Client.Agent.Query().Where(agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))).All(context.Background())
}
