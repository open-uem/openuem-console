package models

import (
	"context"

	"entgo.io/ent/dialect/sql"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/metadata"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) GetMetadataForAgent(agentId string, p partials.PaginationAndSort) ([]*ent.Metadata, error) {
	query := m.Client.Metadata.Query().WithOrg().WithOwner().Where(metadata.HasOwnerWith(agent.ID(agentId)))

	data, err := query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *Model) CountMetadataForAgent(agentId string) (int, error) {
	return m.Client.Metadata.Query().Where(metadata.HasOwnerWith(agent.ID(agentId))).Count(context.Background())
}

func (m *Model) SaveMetadata(agentId string, metadataId int, value string) error {
	return m.Client.Metadata.Create().SetOwnerID(agentId).SetOrgID(metadataId).SetValue(value).OnConflict(sql.ConflictColumns(metadata.OwnerColumn, metadata.OrgColumn)).UpdateNewValues().Exec(context.Background())
}
