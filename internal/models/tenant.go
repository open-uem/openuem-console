package models

import (
	"context"
	"fmt"
	"time"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) CreateDefaultTenant() (*ent.Tenant, error) {
	return m.Client.Tenant.Create().SetDescription("DefaultTenant").SetIsDefault(true).Save(context.Background())
}

func (m *Model) CountTenants() (int, error) {
	return m.Client.Tenant.Query().Count(context.Background())
}

func (m *Model) GetDefaultTenant() (*ent.Tenant, error) {
	return m.Client.Tenant.Query().Where(tenant.IsDefault(true)).Only(context.Background())
}

func (m *Model) GetTenantByID(tenantID int) (*ent.Tenant, error) {
	return m.Client.Tenant.Query().Where(tenant.ID(tenantID)).Only(context.Background())
}

func (m *Model) GetTenants() ([]*ent.Tenant, error) {
	return m.Client.Tenant.Query().All(context.Background())
}

func (m *Model) CountAllTenants(f filters.TenantFilter) (int, error) {
	query := m.Client.Tenant.Query()

	applyTenantsFilter(query, f)

	count, err := query.Count(context.Background())
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *Model) GetTenantsByPage(p partials.PaginationAndSort, f filters.TenantFilter) ([]*ent.Tenant, error) {
	query := m.Client.Tenant.Query()

	applyTenantsFilter(query, f)

	switch p.SortBy {
	case "ID":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(tenant.FieldID))
		} else {
			query.Order(ent.Desc(tenant.FieldID))
		}
	case "name":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(tenant.FieldDescription))
		} else {
			query.Order(ent.Desc(tenant.FieldDescription))
		}
	case "default":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(tenant.FieldIsDefault))
		} else {
			query.Order(ent.Desc(tenant.FieldIsDefault))
		}
	case "created":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(tenant.FieldCreated))
		} else {
			query.Order(ent.Desc(tenant.FieldCreated))
		}
	case "modified":
		if p.SortOrder == "asc" {
			query.Order(ent.Asc(tenant.FieldModified))
		} else {
			query.Order(ent.Desc(tenant.FieldModified))
		}

	default:
		query.Order(ent.Asc(tenant.FieldID))
	}

	return query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
}

func (m *Model) UpdateTenant(tenantID int, desc string, isDefault bool) error {

	query := m.Client.Tenant.Update().Where(tenant.ID(tenantID)).SetDescription(desc)

	if isDefault {
		if err := m.Client.Tenant.Update().Where(tenant.Not(tenant.ID(tenantID))).SetIsDefault(false).Exec(context.Background()); err != nil {
			return err
		}
		return query.SetIsDefault(true).Exec(context.Background())
	} else {
		count, err := m.Client.Tenant.Query().Where(tenant.Not(tenant.ID(tenantID)), tenant.IsDefault(true)).Count(context.Background())
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("this is the current default organization, you cannot remove it as default org until you select a new default org first")
		}
		return query.SetIsDefault(false).Exec(context.Background())
	}
}

func (m *Model) AddTenant(name string, isDefault bool, siteName string) error {
	if isDefault {
		// Remove the is default property for existing orgs
		if err := m.Client.Tenant.Update().SetIsDefault(false).Exec(context.Background()); err != nil {
			return err
		}
	}

	t, err := m.Client.Tenant.Create().SetDescription(name).SetIsDefault(isDefault).Save(context.Background())
	if err != nil {
		return err
	}

	// Clone global settings
	cloneErr := m.CloneGlobalSettings(t.ID)
	if cloneErr != nil {
		// delete tenant as rollback
		if err := m.DeleteTenant(t.ID); err != nil {
			return err
		}
		return cloneErr
	}

	return m.Client.Site.Create().SetDescription(siteName).SetIsDefault(true).SetTenantID(t.ID).Exec(context.Background())
}

func (m *Model) DeleteTenant(tenantID int) error {
	_, err := m.Client.Tenant.Delete().Where(tenant.ID(tenantID)).Exec(context.Background())
	return err
}

func (m *Model) TenantNameTaken(desc string) (bool, error) {
	return m.Client.Tenant.Query().Where(tenant.Description(desc)).Exist(context.Background())
}

func applyTenantsFilter(query *ent.TenantQuery, f filters.TenantFilter) {
	if len(f.Name) > 0 {
		query.Where(tenant.DescriptionContainsFold(f.Name))
	}

	if len(f.CreatedFrom) > 0 {
		dateFrom, err := time.Parse("2006-01-02", f.CreatedFrom)
		if err == nil {
			query.Where(tenant.CreatedGTE(dateFrom))
		}
	}

	if len(f.CreatedTo) > 0 {
		dateTo, err := time.Parse("2006-01-02", f.CreatedTo)
		if err == nil {
			query.Where(tenant.CreatedLTE(dateTo))
		}
	}

	if len(f.ModifiedFrom) > 0 {
		dateFrom, err := time.Parse("2006-01-02", f.ModifiedFrom)
		if err == nil {
			query.Where(tenant.ModifiedGTE(dateFrom))
		}
	}

	if len(f.ModifiedTo) > 0 {
		dateTo, err := time.Parse("2006-01-02", f.ModifiedTo)
		if err == nil {
			query.Where(tenant.ModifiedLTE(dateTo))
		}
	}

	if len(f.DefaultOptions) > 0 {
		if len(f.DefaultOptions) == 1 && f.DefaultOptions[0] == "Yes" {
			query.Where(tenant.IsDefaultEQ(true))
		}

		if len(f.DefaultOptions) == 1 && f.DefaultOptions[0] == "No" {
			query.Where(tenant.IsDefaultEQ(false))
		}
	}
}

func (m *Model) GetAgentsByTenant(tenantID int) ([]*ent.Agent, error) {
	return m.Client.Agent.Query().Where(agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))).All(context.Background())
}
