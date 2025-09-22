package models

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OperatingSystemsTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	p          partials.PaginationAndSort
	commonInfo *partials.CommonInfo
}

func (suite *OperatingSystemsTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	t, err := suite.model.CreateDefaultTenant()
	assert.NoError(suite.T(), err, "should create default tenant")

	s, err := suite.model.CreateDefaultSite(t)
	assert.NoError(suite.T(), err, "should create default site")

	suite.commonInfo = &partials.CommonInfo{TenantID: strconv.Itoa(t.ID), SiteID: strconv.Itoa(s.ID)}

	for i := 0; i <= 6; i++ {
		_, err := client.Agent.Create().
			SetID(fmt.Sprintf("agent%d", i)).
			SetHostname(fmt.Sprintf("agent%d", i)).
			SetOs("windows").
			SetNickname(fmt.Sprintf("agent%d", i)).
			SetAgentStatus(agent.AgentStatusEnabled).
			AddSiteIDs(s.ID).
			Save(context.Background())
		assert.NoError(suite.T(), err, "should create agent")
	}

	for i := 0; i <= 6; i++ {
		err := client.OperatingSystem.Create().
			SetType("windows").
			SetUsername(fmt.Sprintf("user%d", i)).
			SetVersion(fmt.Sprintf("windows%d", i)).
			SetDescription(fmt.Sprintf("description%d", i)).
			SetOwnerID(fmt.Sprintf("agent%d", i)).
			Exec(context.Background())
		assert.NoError(suite.T(), err, "should create operating system")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *OperatingSystemsTestSuite) TestCountAgentsByOSVersion() {
	agents, err := suite.model.CountAgentsByOSVersion(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents by os version")
	assert.Equal(suite.T(), 7, len(agents), "should count 7 agents by os versions")
}

func (suite *OperatingSystemsTestSuite) TestCountAllOSUsernames() {
	count, err := suite.model.CountAllOSUsernames(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all usernames")
	assert.Equal(suite.T(), 7, count, "should count 7 usernames")
}

func (suite *OperatingSystemsTestSuite) TestGetOSVersions() {
	allVersions := []string{"windows0", "windows1", "windows2", "windows3", "windows4", "windows5", "windows6"}
	items, err := suite.model.GetOSVersions(filters.AgentFilter{AgentOSVersions: []string{"windows"}}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get os versions")
	assert.Equal(suite.T(), allVersions, items, "should get 7 os versions")
}

func TestOperatingSystemsTestSuite(t *testing.T) {
	suite.Run(t, new(OperatingSystemsTestSuite))
}
