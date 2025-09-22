package models

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UpdatesTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	p          partials.PaginationAndSort
	commonInfo *partials.CommonInfo
}

func (suite *UpdatesTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	t, err := suite.model.CreateDefaultTenant()
	assert.NoError(suite.T(), err, "should create default tenant")

	s, err := suite.model.CreateDefaultSite(t)
	assert.NoError(suite.T(), err, "should create default site")

	suite.commonInfo = &partials.CommonInfo{TenantID: strconv.Itoa(t.ID), SiteID: strconv.Itoa(s.ID)}

	for i := 0; i <= 6; i++ {
		err := client.Agent.Create().
			SetID(fmt.Sprintf("agent%d", i)).
			SetHostname(fmt.Sprintf("agent%d", i)).
			SetOs("windows").
			SetNickname(fmt.Sprintf("agent%d", i)).
			SetAgentStatus(agent.AgentStatusEnabled).
			AddSiteIDs(s.ID).
			Exec(context.Background())
		assert.NoError(suite.T(), err, "should create agent")
	}

	for i := 0; i <= 6; i++ {
		query := client.SystemUpdate.Create().
			SetOwnerID(fmt.Sprintf("agent%d", i))

		if i%3 == 0 {
			query.SetSystemUpdateStatus(openuem_nats.DISABLED)
		} else {
			query.SetSystemUpdateStatus(openuem_nats.NOTIFY_SCHEDULED_INSTALLATION)
		}

		if i%2 == 0 {
			query.SetPendingUpdates(true)
		} else {
			query.SetPendingUpdates(false)
		}

		query.SetLastInstall(time.Now()).SetLastSearch(time.Now())
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create system update")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *UpdatesTestSuite) TestUpdatesTestSuite() {
	count, err := suite.model.CountAllSystemUpdates(filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f := filters.SystemUpdatesFilter{Nickname: "agent"}
	count, err = suite.model.CountAllSystemUpdates(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f = filters.SystemUpdatesFilter{AgentOSVersions: []string{"windows"}}
	count, err = suite.model.CountAllSystemUpdates(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f = filters.SystemUpdatesFilter{LastSearchFrom: "2024-01-01", LastSearchTo: "2034-01-01"}
	count, err = suite.model.CountAllSystemUpdates(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f = filters.SystemUpdatesFilter{LastInstallFrom: "2024-01-01", LastInstallTo: "2034-01-01"}
	count, err = suite.model.CountAllSystemUpdates(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f = filters.SystemUpdatesFilter{PendingUpdateOptions: []string{"Yes"}}
	count, err = suite.model.CountAllSystemUpdates(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 4, count, "should get 4 updates")

	f = filters.SystemUpdatesFilter{PendingUpdateOptions: []string{"No"}}
	count, err = suite.model.CountAllSystemUpdates(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 3, count, "should get 3 updates")

	f = filters.SystemUpdatesFilter{UpdateStatus: []string{openuem_nats.NOTIFY_SCHEDULED_INSTALLATION}}
	count, err = suite.model.CountAllSystemUpdates(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 4, count, "should get 4 updates")
}

func (suite *UpdatesTestSuite) TestGetSystemUpdatesByPage() {
	items, err := suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Nickname)
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Nickname)
	}

	suite.p.SortBy = "agentOS"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
	}

	suite.p.SortBy = "agentOS"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
	}

	suite.p.SortBy = "updateStatus"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[0].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 3), items[1].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6), items[2].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[3].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[4].Nickname)

	suite.p.SortBy = "updateStatus"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get antiviri by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[0].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[1].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 4), items[2].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 5), items[3].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[4].Nickname)

	suite.p.SortBy = "lastSearch"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
	}

	suite.p.SortBy = "lastSearch"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Nickname)
	}

	suite.p.SortBy = "lastInstall"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
	}

	suite.p.SortBy = "lastInstall"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Nickname)
	}

	suite.p.SortBy = "pendingUpdates"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[0].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 3), items[1].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 5), items[2].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[3].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[4].Nickname)

	suite.p.SortBy = "pendingUpdates"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get system updates by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[0].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[1].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 4), items[2].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6), items[3].Nickname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[4].Nickname)
}

func TestUpdatesTestSuite(t *testing.T) {
	suite.Run(t, new(UpdatesTestSuite))
}
