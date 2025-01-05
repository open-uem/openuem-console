package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem_ent/agent"
	"github.com/open-uem/openuem_ent/enttest"
	"github.com/open-uem/openuem_nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UpdatesTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *UpdatesTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	for i := 0; i <= 6; i++ {
		err := client.Agent.Create().
			SetID(fmt.Sprintf("agent%d", i)).
			SetOs("windows").
			SetHostname(fmt.Sprintf("agent%d", i)).
			SetAgentStatus(agent.AgentStatusEnabled).
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
	count, err := suite.model.CountAllSystemUpdates(filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f := filters.SystemUpdatesFilter{Hostname: "agent"}
	count, err = suite.model.CountAllSystemUpdates(f)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f = filters.SystemUpdatesFilter{AgentOSVersions: []string{"windows"}}
	count, err = suite.model.CountAllSystemUpdates(f)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f = filters.SystemUpdatesFilter{LastSearchFrom: "2024-01-01", LastSearchTo: "2034-01-01"}
	count, err = suite.model.CountAllSystemUpdates(f)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f = filters.SystemUpdatesFilter{LastInstallFrom: "2024-01-01", LastInstallTo: "2034-01-01"}
	count, err = suite.model.CountAllSystemUpdates(f)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 7, count, "should get 7 updates")

	f = filters.SystemUpdatesFilter{PendingUpdateOptions: []string{"Yes"}}
	count, err = suite.model.CountAllSystemUpdates(f)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 4, count, "should get 4 updates")

	f = filters.SystemUpdatesFilter{PendingUpdateOptions: []string{"No"}}
	count, err = suite.model.CountAllSystemUpdates(f)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 3, count, "should get 3 updates")

	f = filters.SystemUpdatesFilter{UpdateStatus: []string{openuem_nats.NOTIFY_SCHEDULED_INSTALLATION}}
	count, err = suite.model.CountAllSystemUpdates(f)
	assert.NoError(suite.T(), err, "should count all system updates")
	assert.Equal(suite.T(), 4, count, "should get 4 updates")
}

func (suite *UpdatesTestSuite) TestGetSystemUpdatesByPage() {
	items, err := suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Hostname)
	}

	suite.p.SortBy = "hostname"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "hostname"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Hostname)
	}

	suite.p.SortBy = "agentOS"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "agentOS"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "updateStatus"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[0].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 3), items[1].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6), items[2].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[3].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[4].Hostname)

	suite.p.SortBy = "updateStatus"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[0].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[1].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 4), items[2].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 5), items[3].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[4].Hostname)

	suite.p.SortBy = "lastSearch"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "lastSearch"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Hostname)
	}

	suite.p.SortBy = "lastInstall"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "lastInstall"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Hostname)
	}

	suite.p.SortBy = "pendingUpdates"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[0].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 3), items[1].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 5), items[2].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[3].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[4].Hostname)

	suite.p.SortBy = "pendingUpdates"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetSystemUpdatesByPage(suite.p, filters.SystemUpdatesFilter{})
	assert.NoError(suite.T(), err, "should get system updates by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[0].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[1].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 4), items[2].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6), items[3].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[4].Hostname)
}

func TestUpdatesTestSuite(t *testing.T) {
	suite.Run(t, new(UpdatesTestSuite))
}
