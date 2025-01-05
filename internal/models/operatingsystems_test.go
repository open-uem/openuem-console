package models

import (
	"context"
	"fmt"
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
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *OperatingSystemsTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	for i := 0; i <= 6; i++ {
		_, err := client.Agent.Create().
			SetID(fmt.Sprintf("agent%d", i)).
			SetOs("windows").
			SetHostname(fmt.Sprintf("agent%d", i)).
			SetAgentStatus(agent.AgentStatusEnabled).
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

func (suite *ComputersTestSuite) TestCountAgentsByOSVersion() {
	agents, err := suite.model.CountAgentsByOSVersion()
	assert.NoError(suite.T(), err, "should count all agents by os version")
	assert.Equal(suite.T(), 7, len(agents), "should count 7 agents by os versions")
}

func (suite *ComputersTestSuite) TestCountAllOSUsernames() {
	count, err := suite.model.CountAllOSUsernames()
	assert.NoError(suite.T(), err, "should count all usernames")
	assert.Equal(suite.T(), 7, count, "should count 7 usernames")
}

func (suite *ComputersTestSuite) TestGetOSVersions() {
	allVersions := []string{"windows0", "windows1", "windows2", "windows3", "windows4", "windows5", "windows6"}
	items, err := suite.model.GetOSVersions(filters.AgentFilter{AgentOSVersions: []string{"windows"}})
	assert.NoError(suite.T(), err, "should get os versions")
	assert.Equal(suite.T(), allVersions, items, "should get 7 os versions")
}

func TestOperatingSystemsTestSuite(t *testing.T) {
	suite.Run(t, new(OperatingSystemsTestSuite))
}
