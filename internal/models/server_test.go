package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem_ent"
	"github.com/open-uem/openuem_ent/enttest"
	"github.com/open-uem/openuem_ent/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServersTestSuite struct {
	suite.Suite
	t        enttest.TestingT
	model    Model
	p        partials.PaginationAndSort
	serverId int
}

func (suite *ServersTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	for i := 0; i <= 6; i++ {
		server, err := client.Server.Create().
			SetArch("amd64").
			SetChannel("stable").
			SetHostname(fmt.Sprintf("server%d", i)).
			SetNatsComponent(true).
			SetNotificationWorkerComponent(true).
			SetConsoleComponent(true).
			SetOs("linux").
			SetVersion(fmt.Sprintf("0.1.%d", i)).
			SetUpdateStatus(server.UpdateStatusPending).
			SetUpdateWhen(time.Now()).
			SetUpdateMessage(fmt.Sprintf("test%d", i)).
			Save(context.Background())
		suite.serverId = server.ID
		assert.NoError(suite.T(), err)
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *ServersTestSuite) TestCountAllUpdateServers() {
	count, err := suite.model.CountAllUpdateServers(filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 7, count, "should count 7 servers")

	f := filters.UpdateServersFilter{
		Hostname: "server",
	}
	count, err = suite.model.CountAllUpdateServers(f)
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 7, count, "should count 7 servers")

	f = filters.UpdateServersFilter{
		UpdateWhenFrom: "2024-12-01",
		UpdateWhenTo:   "2035-01-01",
	}
	count, err = suite.model.CountAllUpdateServers(f)
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 7, count, "should count 7 servers")

	f = filters.UpdateServersFilter{
		UpdateStatus: []string{"Pending"},
	}
	count, err = suite.model.CountAllUpdateServers(f)
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 7, count, "should count 7 servers")

	f = filters.UpdateServersFilter{
		UpdateStatus: []string{"Success"},
	}
	count, err = suite.model.CountAllUpdateServers(f)
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 0, count, "should count 0 servers")

	f = filters.UpdateServersFilter{
		UpdateStatus: []string{"In Progress"},
	}
	count, err = suite.model.CountAllUpdateServers(f)
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 0, count, "should count 0 servers")

	f = filters.UpdateServersFilter{
		UpdateStatus: []string{"Error"},
	}
	count, err = suite.model.CountAllUpdateServers(f)
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 0, count, "should count 0 servers")

	f = filters.UpdateServersFilter{
		Releases: []string{"0.1.1", "0.1.2"},
	}
	count, err = suite.model.CountAllUpdateServers(f)
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 2, count, "should count 2 servers")
}

func (suite *ServersTestSuite) TestGetUpdateServersByPage() {
	servers, err := suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("test%d", 6-i), s.UpdateMessage, fmt.Sprintf("server update message should be test%d for default query", 6-i))
	}

	suite.p.SortBy = "hostname"
	suite.p.SortOrder = "desc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("server%d", 6-i), s.Hostname, fmt.Sprintf("server hostname should be server%d", 6-i))
	}

	suite.p.SortBy = "hostname"
	suite.p.SortOrder = "asc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("server%d", i), s.Hostname, fmt.Sprintf("server hostname should be server%d", i))
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "desc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("0.1.%d", 6-i), s.Version, fmt.Sprintf("server version should be 0.1.%d", 6-i))
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "asc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("0.1.%d", i), s.Version, fmt.Sprintf("server version should be 0.1.%d", i))
	}

	suite.p.SortBy = "status"
	suite.p.SortOrder = "desc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("server%d", i), s.Hostname, fmt.Sprintf("server hostname should be server%d", 6-i))
	}

	suite.p.SortBy = "status"
	suite.p.SortOrder = "asc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("server%d", i), s.Hostname, fmt.Sprintf("server hostname should be server%d", 6-i))
	}

	suite.p.SortBy = "message"
	suite.p.SortOrder = "desc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("test%d", 6-i), s.UpdateMessage, fmt.Sprintf("server update message should be test%d", 6-i))
	}

	suite.p.SortBy = "message"
	suite.p.SortOrder = "asc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("test%d", i), s.UpdateMessage, fmt.Sprintf("server update message should be test%d", i))
	}

	suite.p.SortBy = "when"
	suite.p.SortOrder = "desc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("test%d", 6-i), s.UpdateMessage, fmt.Sprintf("server update message should be test%d", 6-i))
	}

	suite.p.SortBy = "when"
	suite.p.SortOrder = "asc"
	servers, err = suite.model.GetUpdateServersByPage(suite.p, filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get update servers by page")
	for i, s := range servers {
		assert.Equal(suite.T(), fmt.Sprintf("test%d", i), s.UpdateMessage, fmt.Sprintf("server update message should be test%d", i))
	}
}

func (suite *ServersTestSuite) TestGetHigherServerReleaseInstalled() {
	server, err := suite.model.GetHigherServerReleaseInstalled()
	assert.NoError(suite.T(), err, "should get higher server release")
	assert.Equal(suite.T(), "0.1.6", server.Version, "higher server release should be 0.1.6")
}

func (suite *ServersTestSuite) TestGetAppliedReleases() {
	releases, err := suite.model.GetAppliedReleases()
	assert.NoError(suite.T(), err, "should get applied releases")
	assert.Equal(suite.T(), []string{"0.1.6", "0.1.5", "0.1.4", "0.1.3", "0.1.2", "0.1.1", "0.1.0"}, releases, "server releases applied should be ['0.1.6','0.1.5','0.1.4','0.1.3','0.1.2','0.1.1','0.1.0']")
}

func (suite *ServersTestSuite) TestSaveServerUpdateInfo() {
	err := suite.model.SaveServerUpdateInfo(suite.serverId, "Success", "installed", "0.2.0")
	assert.NoError(suite.T(), err, "should save server info")

	s, err := suite.model.GetServerById(suite.serverId)
	assert.NoError(suite.T(), err, "should get server info")
	assert.Equal(suite.T(), "server6", s.Hostname)
	assert.Equal(suite.T(), server.UpdateStatusSuccess, s.UpdateStatus)
	assert.Equal(suite.T(), "installed", s.UpdateMessage)
	assert.Equal(suite.T(), "0.2.0", s.Version)
}

func (suite *ServersTestSuite) TestGetAllUpdateServers() {
	servers, err := suite.model.GetAllUpdateServers(filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should get all update servers")
	assert.Equal(suite.T(), 7, len(servers), "should get 7 servers")

	servers, err = suite.model.GetAllUpdateServers(filters.UpdateServersFilter{Hostname: "server"})
	assert.NoError(suite.T(), err, "should get all update servers")
	assert.Equal(suite.T(), 7, len(servers), "should get 7 servers")
}

func (suite *ServersTestSuite) TestGetServerById() {
	server, err := suite.model.GetServerById(suite.serverId)
	assert.NoError(suite.T(), err, "should get server by id")
	assert.Equal(suite.T(), "server6", server.Hostname, "server should have server6 hostname")

	server, err = suite.model.GetServerById(0)
	assert.Error(suite.T(), err, "should get an error using a non existent id")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "query should return a not found error")
}

func (suite *ServersTestSuite) TestDeleteServer() {
	err := suite.model.DeleteServer(suite.serverId)
	assert.NoError(suite.T(), err, "should delete server")

	count, err := suite.model.CountAllUpdateServers(filters.UpdateServersFilter{})
	assert.NoError(suite.T(), err, "should count all update servers")
	assert.Equal(suite.T(), 6, count, "should count 6 servers")
}

func (suite *ServersTestSuite) TestServersExists() {
	exists, err := suite.model.ServersExists()
	assert.NoError(suite.T(), err, "servers should exist")
	assert.Equal(suite.T(), true, exists, "servers should exist again")
}

func TestServersTestSuite(t *testing.T) {
	suite.Run(t, new(ServersTestSuite))
}
