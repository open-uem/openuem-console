package models

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AppsTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	p          partials.PaginationAndSort
	commonInfo *partials.CommonInfo
}

func (suite *AppsTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	t, err := suite.model.CreateDefaultTenant()
	assert.NoError(suite.T(), err, "should create default tenant")

	s, err := suite.model.CreateDefaultSite(t)
	assert.NoError(suite.T(), err, "should create default site")

	suite.commonInfo = &partials.CommonInfo{TenantID: strconv.Itoa(t.ID), SiteID: strconv.Itoa(s.ID)}

	err = client.Agent.Create().
		SetID("agent1").
		SetHostname("agent1").
		SetOs("windows").
		SetNickname("agent1").
		SetAgentStatus(agent.AgentStatusEnabled).
		AddSiteIDs(s.ID).
		Exec(context.Background())
	assert.NoError(suite.T(), err, "should create agent")

	for i := 0; i <= 6; i++ {
		err := client.App.Create().
			SetName(fmt.Sprintf("app%d", i)).
			SetPublisher(fmt.Sprintf("publisher%d", i)).
			SetVersion(fmt.Sprintf("version%d", i)).
			SetInstallDate(time.Now().Format("2006-01-02")).
			SetOwnerID("agent1").
			Exec(context.Background())
		assert.NoError(suite.T(), err)
	}
}

func (suite *AppsTestSuite) TestCountAgentApps() {

	f := filters.ApplicationsFilter{}

	count, err := suite.model.CountAgentApps("agent1", f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count agent apps")
	assert.Equal(suite.T(), 7, count, "should count 7 apps")

	count, err = suite.model.CountAgentApps("agent9", f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count agent apps")
	assert.Equal(suite.T(), 0, count, "should count 0 apps")
}

func (suite *AppsTestSuite) TestCountAllApps() {
	count, err := suite.model.CountAllApps(filters.ApplicationsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all apps")
	assert.Equal(suite.T(), 7, count, "should count 7 apps")

	count, err = suite.model.CountAllApps(filters.ApplicationsFilter{AppName: "app5"}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all apps")
	assert.Equal(suite.T(), 1, count, "should count 1 apps")

	count, err = suite.model.CountAllApps(filters.ApplicationsFilter{Vendor: "publisher"}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all apps")
	assert.Equal(suite.T(), 7, count, "should count 7 apps")
}

func (suite *AppsTestSuite) TestGetAgentAppsByPage() {
	f := filters.ApplicationsFilter{}

	suite.p.SortBy = "name"
	suite.p.SortOrder = "asc"
	items, err := suite.model.GetAgentAppsByPage("agent1", suite.p, f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agent apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("app%d", i), item.Name)
	}

	suite.p.SortBy = "name"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentAppsByPage("agent1", suite.p, f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agent apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("app%d", 6-i), item.Name)
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentAppsByPage("agent1", suite.p, f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agent apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("version%d", i), item.Version)
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentAppsByPage("agent1", suite.p, f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agent apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("version%d", 6-i), item.Version)
	}

	suite.p.SortBy = "publisher"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentAppsByPage("agent1", suite.p, f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agent apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("publisher%d", i), item.Publisher)
	}

	suite.p.SortBy = "publisher"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentAppsByPage("agent1", suite.p, f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agent apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("publisher%d", 6-i), item.Publisher)
	}

	suite.p.SortBy = "installation"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentAppsByPage("agent1", suite.p, f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agent apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("publisher%d", i), item.Publisher)
	}

	suite.p.SortBy = "installation"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentAppsByPage("agent1", suite.p, f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agent apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("publisher%d", i), item.Publisher)
	}
}

func (suite *AppsTestSuite) TestGetAppsByPage() {
	suite.p.SortBy = "name"
	suite.p.SortOrder = "asc"
	items, err := suite.model.GetAppsByPage(suite.p, filters.ApplicationsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get apps by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("app%d", i), item.Name)
	}

	suite.p.SortBy = "name"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAppsByPage(suite.p, filters.ApplicationsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get apps by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("app%d", 6-i), item.Name)
	}

	suite.p.SortBy = "publisher"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAppsByPage(suite.p, filters.ApplicationsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get apps by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("publisher%d", i), item.Publisher)
	}

	suite.p.SortBy = "publisher"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAppsByPage(suite.p, filters.ApplicationsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get apps by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("publisher%d", 6-i), item.Publisher)
	}

	suite.p.SortBy = "installations"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAppsByPage(suite.p, filters.ApplicationsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get apps by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("publisher%d", i), item.Publisher)
	}

	suite.p.SortBy = "installations"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAppsByPage(suite.p, filters.ApplicationsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get apps by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("publisher%d", i), item.Publisher)
	}
}

func (suite *AppsTestSuite) TestGetTop10InstalledApps() {
	items, err := suite.model.GetTop10InstalledApps()
	assert.NoError(suite.T(), err, "should get top 10 installed apps")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("app%d", 6-i), item.Name)
		assert.Equal(suite.T(), 1, item.Count)
	}
}

func TestAppsTestSuite(t *testing.T) {
	suite.Run(t, new(AppsTestSuite))
}
