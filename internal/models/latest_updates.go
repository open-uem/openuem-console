package models

import (
	"context"

	"github.com/open-uem/openuem-console/internal/views/partials"
	ent "github.com/open-uem/openuem_ent"
	"github.com/open-uem/openuem_ent/agent"
	"github.com/open-uem/openuem_ent/update"
)

func (m *Model) CountLatestUpdates(agentId string) (int, error) {
	return m.Client.Update.Query().Where(update.HasOwnerWith(agent.ID(agentId))).Count(context.Background())
}

func (m *Model) GetLatestUpdates(agentId string, p partials.PaginationAndSort) ([]*ent.Update, error) {
	query := m.Client.Update.Query().Where(update.HasOwnerWith(agent.ID(agentId)))

	switch p.SortBy {
	case "title":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(update.FieldTitle))
		} else {
			query = query.Order(ent.Desc(update.FieldTitle))
		}
	case "date":
		if p.SortOrder == "asc" {
			query = query.Order(ent.Asc(update.FieldDate))
		} else {
			query = query.Order(ent.Desc(update.FieldDate))
		}
	default:
		query = query.Order(ent.Desc(update.FieldDate))
	}

	updates, err := query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
	if err != nil {
		return nil, err
	}

	return updates, nil
}
