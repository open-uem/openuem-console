package models

import (
	"context"
	"strconv"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/orgmetadata"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) GetAllOrgMetadata(c *partials.CommonInfo) ([]*ent.OrgMetadata, error) {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	data, err := m.Client.OrgMetadata.Query().Where(orgmetadata.HasTenantWith(tenant.ID(tenantID))).All(context.Background())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *Model) GetOrgMetadataByPage(p partials.PaginationAndSort, c *partials.CommonInfo) ([]*ent.OrgMetadata, error) {
	var err error
	var data []*ent.OrgMetadata

	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	query := m.Client.OrgMetadata.Query().Where(orgmetadata.HasTenantWith(tenant.ID(tenantID))).Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	switch p.SortBy {
	case "name":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(orgmetadata.FieldName))
		} else {
			query = query.Order(ent.Desc(orgmetadata.FieldName))
		}
	case "description":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(orgmetadata.FieldDescription))
		} else {
			query = query.Order(ent.Desc(orgmetadata.FieldDescription))
		}
	default:
		query = query.Order(ent.Asc(orgmetadata.FieldID))
	}

	data, err = query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *Model) CountAllOrgMetadata(c *partials.CommonInfo) (int, error) {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return -1, err
	}

	return m.Client.OrgMetadata.Query().Where(orgmetadata.HasTenantWith(tenant.ID(tenantID))).Count(context.Background())
}

func (m *Model) NewOrgMetadata(name, description string, c *partials.CommonInfo) error {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	return m.Client.OrgMetadata.Create().SetName(name).SetDescription(description).SetTenantID(tenantID).Exec(context.Background())
}

func (m *Model) UpdateOrgMetadata(id int, name, description string, c *partials.CommonInfo) error {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	return m.Client.OrgMetadata.Update().SetName(name).SetDescription(description).Where(orgmetadata.ID(id), orgmetadata.HasTenantWith(tenant.ID(tenantID))).Exec(context.Background())
}

func (m *Model) DeleteOrgMetadata(id int, c *partials.CommonInfo) error {
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return err
	}

	return m.Client.OrgMetadata.DeleteOneID(id).Where(orgmetadata.HasTenantWith(tenant.ID(tenantID))).Exec(context.Background())
}
