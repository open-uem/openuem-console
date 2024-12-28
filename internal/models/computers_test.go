package models

import (
	"context"
	"fmt"
	"testing"

	"github.com/doncicuto/openuem-console/internal/views/filters"
	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/agent"
	"github.com/doncicuto/openuem_ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ComputersTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *ComputersTestSuite) SetupTest() {
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

	for i := 0; i <= 6; i++ {
		query := client.Computer.Create().
			SetManufacturer(fmt.Sprintf("manufacturer%d", i)).
			SetMemory(10240000000).
			SetModel(fmt.Sprintf("model%d", i)).
			SetProcessor("intel").
			SetProcessorArch("amd64").
			SetProcessorCores(4).
			SetOwnerID(fmt.Sprintf("agent%d", i))
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create computer")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *ComputersTestSuite) TestCountAllComputers() {
	count, err := suite.model.CountAllComputers(filters.AgentFilter{})
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 7, count, "should count 7 computers")

	f := filters.AgentFilter{Hostname: "agent"}
	count, err = suite.model.CountAllComputers(f)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 7, count, "should count 7 computers")

	f = filters.AgentFilter{AgentOSVersions: []string{"windows"}}
	count, err = suite.model.CountAllComputers(f)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 7, count, "should count 7 computers")

	f = filters.AgentFilter{ComputerManufacturers: []string{"manufacturer0", "manufacturer1", "manufacturer3"}}
	count, err = suite.model.CountAllComputers(f)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 3, count, "should count 3 computers")

	f = filters.AgentFilter{ComputerModels: []string{"model0", "model1", "model2", "model3"}}
	count, err = suite.model.CountAllComputers(f)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 4, count, "should count 4 computers")

	f = filters.AgentFilter{OSVersions: []string{"windows1", "windows4"}}
	count, err = suite.model.CountAllComputers(f)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 2, count, "should count 2 computers")

	f = filters.AgentFilter{Username: "user1"}
	count, err = suite.model.CountAllComputers(f)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 1, count, "should count 1 computers")
}

func (suite *ComputersTestSuite) TestGetAgentComputerInfo() {
	item, err := suite.model.GetAgentComputerInfo("agent1")
	assert.NoError(suite.T(), err, "should found agent1")
	assert.Equal(suite.T(), "manufacturer1", item.Edges.Computer.Manufacturer, "manufacturer should be manufacturer1")

	item, err = suite.model.GetAgentComputerInfo("agent7")
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestGetAgentOSInfo() {
	item, err := suite.model.GetAgentOSInfo("agent3")
	assert.NoError(suite.T(), err, "should found agent3")
	assert.Equal(suite.T(), "windows3", item.Edges.Operatingsystem.Version, "version should be windows3")
	assert.Equal(suite.T(), "user3", item.Edges.Operatingsystem.Username, "user should be user3")

	item, err = suite.model.GetAgentOSInfo("agent7")
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestGetComputerManufacturers() {
	allManufacturers := []string{"manufacturer0", "manufacturer1", "manufacturer2", "manufacturer3", "manufacturer4", "manufacturer5", "manufacturer6"}
	items, err := suite.model.GetComputerManufacturers()
	assert.NoError(suite.T(), err, "should get computer manufacturers")
	assert.Equal(suite.T(), 7, len(allManufacturers), "should get 7 manufacturers")
	assert.Equal(suite.T(), allManufacturers, items, "should get 7 manufacturers")
}

func (suite *ComputersTestSuite) TestGetComputerModels() {
	allModels := []string{"model1", "model2"}
	items, err := suite.model.GetComputerModels(filters.AgentFilter{ComputerManufacturers: []string{"manufacturer1", "manufacturer2"}})
	assert.NoError(suite.T(), err, "should get computer models")
	assert.Equal(suite.T(), allModels, items, "should get two computer models")
}

func (suite *ComputersTestSuite) TestGetOSVersions() {
	allVersions := []string{"windows0", "windows1", "windows2", "windows3", "windows4", "windows5", "windows6"}
	items, err := suite.model.GetOSVersions(filters.AgentFilter{AgentOSVersions: []string{"windows"}})
	assert.NoError(suite.T(), err, "should get os versions")
	assert.Equal(suite.T(), allVersions, items, "should get 7 os versions")
}

func (suite *ComputersTestSuite) TestCountAllOSUsernames() {
	count, err := suite.model.CountAllOSUsernames()
	assert.NoError(suite.T(), err, "should count all usernames")
	assert.Equal(suite.T(), 7, count, "should count 7 usernames")
}

func TestComputersTestSuite(t *testing.T) {
	suite.Run(t, new(ComputersTestSuite))
}
