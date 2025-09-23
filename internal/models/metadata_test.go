package models

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MetadataTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	p          partials.PaginationAndSort
	orgs       []int
	commonInfo *partials.CommonInfo
}

func (suite *MetadataTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	t, err := suite.model.CreateDefaultTenant()
	assert.NoError(suite.T(), err, "should create default tenant")

	s, err := suite.model.CreateDefaultSite(t)
	assert.NoError(suite.T(), err, "should create default site")

	suite.commonInfo = &partials.CommonInfo{TenantID: strconv.Itoa(t.ID), SiteID: strconv.Itoa(s.ID)}

	_, err = client.Agent.Create().
		SetID("agent1").
		SetHostname("agent1").
		SetOs("windows").
		SetNickname("agent1").
		SetAgentStatus(agent.AgentStatusEnabled).
		AddSiteIDs(s.ID).
		Save(context.Background())
	assert.NoError(suite.T(), err, "should create agent")

	orgs := []int{}
	for i := 0; i <= 7; i++ {
		m, err := client.OrgMetadata.Create().
			SetName(fmt.Sprintf("metadata%d", i)).
			SetDescription(fmt.Sprintf("metadata%d description", i)).
			Save(context.Background())
		assert.NoError(suite.T(), err, "set metadata description")
		orgs = append(orgs, m.ID)
	}
	suite.orgs = orgs

	for i := 0; i <= 6; i++ {
		err := client.Metadata.Create().
			SetOrgID(orgs[i]).
			SetOwnerID("agent1").
			SetValue(fmt.Sprintf("value%d", i)).
			Exec(context.Background())
		assert.NoError(suite.T(), err, "should create operating system")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *MetadataTestSuite) TestGetMetadataForAgent() {
	items, err := suite.model.GetMetadataForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get metadata for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("value%d", i), item.Value)
	}
}

func (suite *MetadataTestSuite) TestCountMetadataForAgent() {
	count, err := suite.model.CountMetadataForAgent("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all metadata")
	assert.Equal(suite.T(), 7, count, "should count 7 metadata")
}

func (suite *MetadataTestSuite) TestSaveMetadata() {
	err := suite.model.SaveMetadata("agent1", suite.orgs[7], "value7")
	assert.NoError(suite.T(), err, "should save metadata value")

	count, err := suite.model.CountMetadataForAgent("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all metadata")
	assert.Equal(suite.T(), 8, count, "should count 8 metadata")
}

func TestMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataTestSuite))
}
