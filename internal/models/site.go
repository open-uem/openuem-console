package models

import (
	"context"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
)

func (m *Model) CreateDefaultSite(tenant *ent.Tenant) (*ent.Site, error) {
	return m.Client.Site.Create().SetDescription("DefaultSite").SetIsDefault(true).SetTenantID(tenant.ID).Save(context.Background())
}

func (m *Model) CountSites() (int, error) {
	return m.Client.Site.Query().Count(context.Background())
}

func (m *Model) GetDefaultSite(t *ent.Tenant) (*ent.Site, error) {
	return m.Client.Site.Query().Where(site.IsDefault(true), site.HasTenantWith(tenant.ID(t.ID))).Only(context.Background())
}
