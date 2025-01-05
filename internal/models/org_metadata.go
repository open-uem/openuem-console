package models

import (
	"context"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/orgmetadata"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) GetAllOrgMetadata() ([]*ent.OrgMetadata, error) {
	data, err := m.Client.OrgMetadata.Query().All(context.Background())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *Model) GetOrgMetadataByPage(p partials.PaginationAndSort) ([]*ent.OrgMetadata, error) {
	var err error
	var data []*ent.OrgMetadata

	query := m.Client.OrgMetadata.Query().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

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

func (m *Model) CountAllOrgMetadata() (int, error) {
	return m.Client.OrgMetadata.Query().Count(context.Background())
}

func (m *Model) NewOrgMetadata(name, description string) error {
	return m.Client.OrgMetadata.Create().SetName(name).SetDescription(description).Exec(context.Background())
}

func (m *Model) UpdateOrgMetadata(id int, name, description string) error {
	return m.Client.OrgMetadata.Update().SetName(name).SetDescription(description).Where(orgmetadata.ID(id)).Exec(context.Background())
}

func (m *Model) DeleteOrgMetadata(id int) error {
	return m.Client.OrgMetadata.DeleteOneID(id).Exec(context.Background())
}
