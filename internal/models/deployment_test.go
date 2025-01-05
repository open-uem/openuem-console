package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem_ent/agent"
	"github.com/open-uem/openuem_ent/enttest"
	"github.com/open-uem/openuem_nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DeploymentTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
	orgs  []int
}

func (suite *DeploymentTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	_, err := client.Agent.Create().
		SetID("agent1").
		SetOs("windows").
		SetHostname("agent1").
		SetAgentStatus(agent.AgentStatusEnabled).
		Save(context.Background())
	assert.NoError(suite.T(), err, "should create agent")

	for i := 0; i <= 6; i++ {
		err := client.Deployment.Create().
			SetName(fmt.Sprintf("deployment%d", i)).
			SetVersion(fmt.Sprintf("version%d", i)).
			SetOwnerID("agent1").
			SetPackageID(fmt.Sprintf("package%d", i)).
			SetInstalled(time.Now()).
			SetUpdated(time.Now()).
			Exec(context.Background())
		assert.NoError(suite.T(), err, "should create operating system")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *DeploymentTestSuite) TestGetDeploymentsForAgent() {

	suite.p.SortBy = "name"
	suite.p.SortOrder = "asc"
	items, err := suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", i), item.PackageID)
	}

	suite.p.SortBy = "name"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", 6-i), item.PackageID)
	}

	suite.p.SortBy = "installation"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", i), item.PackageID)
	}

	suite.p.SortBy = "installation"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", 6-i), item.PackageID)
	}

	suite.p.SortBy = "updated"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", i), item.PackageID)
	}

	suite.p.SortBy = "updated"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", 6-i), item.PackageID)
	}
}

func (suite *DeploymentTestSuite) TestDeploymentAlreadyInstalled() {
	installed, err := suite.model.DeploymentAlreadyInstalled("agent1", "package6")
	assert.NoError(suite.T(), err, "should check if deployment already installed")
	assert.Equal(suite.T(), true, installed, "deployment should be installed")

	installed, err = suite.model.DeploymentAlreadyInstalled("agent1", "package7")
	assert.NoError(suite.T(), err, "should check if deployment already installed")
	assert.Equal(suite.T(), false, installed, "deployment should not be installed")
}

func (suite *DeploymentTestSuite) TestCountDeploymentsForAgent() {
	count, err := suite.model.CountDeploymentsForAgent("agent1")
	assert.NoError(suite.T(), err, "should count all deployments for agent")
	assert.Equal(suite.T(), 7, count, "should count 7 deployments for agent")
}

func (suite *DeploymentTestSuite) TestCountAllDeployments() {
	count, err := suite.model.CountAllDeployments()
	assert.NoError(suite.T(), err, "should count all deployments")
	assert.Equal(suite.T(), 7, count, "should count 7 deployments")
}

func (suite *DeploymentTestSuite) TestSaveDeployInfo() {
	err := suite.model.SaveDeployInfo(&openuem_nats.DeployAction{
		AgentId:        "agent1",
		Action:         "install",
		PackageId:      "package7",
		PackageName:    "Package 7",
		PackageVersion: "version7",
		Info:           "info",
	})
	assert.NoError(suite.T(), err, "should save deployment info")

	suite.p.PageSize = 10
	items, err := suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	assert.Equal(suite.T(), "package7", items[7].PackageID, "should get package7")

	err = suite.model.SaveDeployInfo(&openuem_nats.DeployAction{
		AgentId:   "agent1",
		Action:    "update",
		PackageId: "package7",
	})
	assert.NoError(suite.T(), err, "should save deployment info")
	suite.p.PageSize = 10
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	assert.Equal(suite.T(), true, items[7].Updated.IsZero(), "update time should be zero")

	err = suite.model.SaveDeployInfo(&openuem_nats.DeployAction{
		AgentId:   "agent1",
		Action:    "uninstall",
		PackageId: "package3",
	})
	assert.NoError(suite.T(), err, "should save deployment info")
	suite.p.PageSize = 10
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	assert.Equal(suite.T(), true, items[3].Installed.IsZero(), "install time should be zero")
}

func TestDeploymentTestSuite(t *testing.T) {
	suite.Run(t, new(DeploymentTestSuite))
}
