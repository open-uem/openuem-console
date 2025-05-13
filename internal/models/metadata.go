package models

import (
	"context"
	"strconv"

	"entgo.io/ent/dialect/sql"
	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/metadata"
	"github.com/open-uem/ent/site"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (m *Model) GetMetadataForAgent(agentId string, p partials.PaginationAndSort, c *partials.CommonInfo) ([]*ent.Metadata, error) {
	var query *ent.MetadataQuery

	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return nil, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return nil, err
	}

	if siteID == -1 {
		query = m.Client.Metadata.Query().WithOrg().WithOwner().Where(metadata.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID)))))
	} else {
		query = m.Client.Metadata.Query().WithOrg().WithOwner().Where(metadata.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID)))))
	}

	data, err := query.Limit(p.PageSize).Offset((p.CurrentPage - 1) * p.PageSize).All(context.Background())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *Model) CountMetadataForAgent(agentId string, c *partials.CommonInfo) (int, error) {
	siteID, err := strconv.Atoi(c.SiteID)
	if err != nil {
		return 0, err
	}
	tenantID, err := strconv.Atoi(c.TenantID)
	if err != nil {
		return 0, err
	}

	if siteID == -1 {
		return m.Client.Metadata.Query().Where(metadata.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	} else {
		return m.Client.Metadata.Query().Where(metadata.HasOwnerWith(agent.ID(agentId), agent.HasSiteWith(site.ID(siteID), site.HasTenantWith(tenant.ID(tenantID))))).Count(context.Background())
	}
}

func (m *Model) SaveMetadata(agentId string, metadataId int, value string) error {
	return m.Client.Metadata.Create().SetOwnerID(agentId).SetOrgID(metadataId).SetValue(value).OnConflict(sql.ConflictColumns(metadata.OwnerColumn, metadata.OrgColumn)).UpdateNewValues().Exec(context.Background())
}
