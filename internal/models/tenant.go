package models

import (
	"context"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/tenant"
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
