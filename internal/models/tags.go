package models

import (
	"context"

	"github.com/doncicuto/openuem-console/internal/views/partials"
	ent "github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/tag"
)

func (m *Model) GetAllTags() ([]*ent.Tag, error) {
	tags, err := m.Client.Tag.Query().All(context.Background())
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (m *Model) GetTagsByPage(p partials.PaginationAndSort) ([]*ent.Tag, error) {
	var err error
	var tags []*ent.Tag

	query := m.Client.Tag.Query().WithOwner().Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize)

	switch p.SortBy {
	case "tag":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(tag.FieldTag))
		} else {
			query = query.Order(ent.Desc(tag.FieldTag))
		}
	case "description":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(tag.FieldDescription))
		} else {
			query = query.Order(ent.Desc(tag.FieldDescription))
		}
	default:
		query = query.Order(ent.Asc(tag.FieldID))
	}

	tags, err = query.All(context.Background())
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (m *Model) CountAllTags() (int, error) {
	return m.Client.Tag.Query().Count(context.Background())
}

func (m *Model) NewTag(title, description, color string) error {
	return m.Client.Tag.Create().SetTag(title).SetDescription(description).SetColor(color).Exec(context.Background())
}

func (m *Model) UpdateTag(tagId int, title, description, color string) error {
	return m.Client.Tag.Update().SetTag(title).SetDescription(description).SetColor(color).Where(tag.ID(tagId)).Exec(context.Background())
}

func (m *Model) DeleteTag(tagId int) error {
	return m.Client.Tag.DeleteOneID(tagId).Exec(context.Background())
}
