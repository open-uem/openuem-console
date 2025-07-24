package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LatestUpdatesTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *LatestUpdatesTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	_, err := client.Agent.Create().
		SetID("agent1").
		SetOs("windows").
		SetNickname("agent1").
		SetAgentStatus(agent.AgentStatusEnabled).
		Save(context.Background())
	assert.NoError(suite.T(), err, "should create agent")

	for i := 0; i <= 6; i++ {
		err := client.Update.Create().
			SetTitle(fmt.Sprintf("update%d", i)).
			SetDate(time.Now()).
			SetSupportURL("url").
			SetOwnerID("agent1").
			Exec(context.Background())
		assert.NoError(suite.T(), err)
	}
}

func (suite *LatestUpdatesTestSuite) TestCountLatestUpdates() {
	count, err := suite.model.CountLatestUpdates("agent1", &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count lates updates")
	assert.Equal(suite.T(), 7, count, "should have 7 updates")

	count, err = suite.model.CountLatestUpdates("agent2", &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count lates updates")
	assert.Equal(suite.T(), 0, count, "should have 0 updates")
}

func (suite *LatestUpdatesTestSuite) TestGetLatestUpdates() {
	items, err := suite.model.GetLatestUpdates("agent1", suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get latest updates")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("update%d", 6-i), item.Title)
	}

	suite.p.SortBy = "title"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetLatestUpdates("agent1", suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get latest updates")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("update%d", i), item.Title)
	}

	suite.p.SortBy = "title"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetLatestUpdates("agent1", suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get latest updates")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("update%d", 6-i), item.Title)
	}

	suite.p.SortBy = "date"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetLatestUpdates("agent1", suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get latest updates")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("update%d", i), item.Title)
	}

	suite.p.SortBy = "date"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetLatestUpdates("agent1", suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get latest updates")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("update%d", 6-i), item.Title)
	}
}

func TestLatestUpdatesTestSuite(t *testing.T) {
	suite.Run(t, new(LatestUpdatesTestSuite))
}
