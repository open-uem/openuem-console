package models

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/deployment"
	"github.com/open-uem/ent/enttest"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DeploymentTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	p          partials.PaginationAndSort
	commonInfo *partials.CommonInfo
}

func (suite *DeploymentTestSuite) SetupTest() {
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
	items, err := suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", i), item.PackageID)
	}

	suite.p.SortBy = "name"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", 6-i), item.PackageID)
	}

	suite.p.SortBy = "installation"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", i), item.PackageID)
	}

	suite.p.SortBy = "installation"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", 6-i), item.PackageID)
	}

	suite.p.SortBy = "updated"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", i), item.PackageID)
	}

	suite.p.SortBy = "updated"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("package%d", 6-i), item.PackageID)
	}
}

func (suite *DeploymentTestSuite) TestDeploymentAlreadyInstalled() {
	installed, err := suite.model.DeploymentAlreadyInstalled("agent1", "package6", suite.commonInfo)
	assert.NoError(suite.T(), err, "should check if deployment already installed")
	assert.Equal(suite.T(), true, installed, "deployment should be installed")

	installed, err = suite.model.DeploymentAlreadyInstalled("agent1", "package7", suite.commonInfo)
	assert.NoError(suite.T(), err, "should check if deployment already installed")
	assert.Equal(suite.T(), false, installed, "deployment should not be installed")
}

func (suite *DeploymentTestSuite) TestCountDeploymentsForAgent() {
	count, err := suite.model.CountDeploymentsForAgent("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all deployments for agent")
	assert.Equal(suite.T(), 7, count, "should count 7 deployments for agent")
}

func (suite *DeploymentTestSuite) TestCountAllDeployments() {
	count, err := suite.model.CountAllDeployments(suite.commonInfo)
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
	}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should save deployment info")

	suite.p.PageSize = 10
	items, err := suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	assert.Equal(suite.T(), "package7", items[7].PackageID, "should get package7")

	err = suite.model.SaveDeployInfo(&openuem_nats.DeployAction{
		AgentId:   "agent1",
		Action:    "update",
		PackageId: "package7",
	}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should save deployment info")
	suite.p.PageSize = 10
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	assert.Equal(suite.T(), true, items[7].Updated.IsZero(), "update time should be zero")

	err = suite.model.SaveDeployInfo(&openuem_nats.DeployAction{
		AgentId:   "agent1",
		Action:    "uninstall",
		PackageId: "package3",
	}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should save deployment info")
	suite.p.SortBy = "name"
	suite.p.SortOrder = "desc"
	suite.p.PageSize = 10
	items, err = suite.model.GetDeploymentsForAgent("agent1", suite.p, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get deployments for agent")
	assert.Equal(suite.T(), "deployment3", items[3].Name, "install time should be zero")
	assert.Equal(suite.T(), true, items[3].Installed.IsZero(), "install time should be zero")
}

func (suite *DeploymentTestSuite) TestSaveDeployInfoInstallAgain() {
	countDeployments := func(packageID string) int {
		n, err := suite.model.Client.Deployment.Query().Where(deployment.PackageID(packageID), deployment.HasOwnerWith(agent.ID("agent1"))).Count(context.Background())
		assert.NoError(suite.T(), err)
		return n
	}

	// Installing a package that already has a deployment reuses it, as pending
	err := suite.model.SaveDeployInfo(&openuem_nats.DeployAction{AgentId: "agent1", Action: "install", PackageId: "package1", PackageName: "Package 1"}, false, suite.commonInfo)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, countDeployments("package1"), "no second deployment")
	d, err := suite.model.Client.Deployment.Query().Where(deployment.PackageID("package1")).Only(context.Background())
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), d.Installed.IsZero(), "pending again")

	// Duplicates created before are merged, so the agent's result can be saved again
	for i := 0; i < 2; i++ {
		err = suite.model.Client.Deployment.Create().SetName("dup").SetVersion("1").SetOwnerID("agent1").SetPackageID("package2").SetInstalled(time.Now()).SetUpdated(time.Now()).Exec(context.Background())
		assert.NoError(suite.T(), err)
	}
	assert.Equal(suite.T(), 3, countDeployments("package2"))
	err = suite.model.SaveDeployInfo(&openuem_nats.DeployAction{AgentId: "agent1", Action: "install", PackageId: "package2", PackageName: "Package 2"}, false, suite.commonInfo)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, countDeployments("package2"), "duplicates merged")
}

func TestDeploymentTestSuite(t *testing.T) {
	suite.Run(t, new(DeploymentTestSuite))
}

func (suite *DeploymentTestSuite) TestSaveDeployInfoConcurrentInstalls() {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(suite.T(), suite.model.SaveDeployInfo(&openuem_nats.DeployAction{AgentId: "agent1", Action: "install", PackageId: "package.concurrent", PackageName: "Concurrent"}, false, suite.commonInfo))
		}()
	}
	wg.Wait()
	n, err := suite.model.Client.Deployment.Query().Where(deployment.PackageID("package.concurrent")).Count(context.Background())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, n, "one deployment for concurrent requests")
}
