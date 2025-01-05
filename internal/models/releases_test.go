package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem_ent"
	"github.com/open-uem/openuem_ent/agent"
	"github.com/open-uem/openuem_ent/enttest"
	"github.com/open-uem/openuem_ent/release"
	"github.com/open-uem/openuem_nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ReleasesTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *ReleasesTestSuite) SetupTest() {
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
		query := client.Release.Create().
			SetArch("amd64").
			SetChannel("stable").
			SetChecksum("checksum").
			SetFileURL("fileurl").
			SetIsCritical(false).
			SetReleaseDate(time.Now()).
			SetReleaseNotes("url").
			SetVersion(fmt.Sprintf("0.1.%d", i)).
			AddAgentIDs(fmt.Sprintf("agent%d", i))

		if i%2 == 0 {
			query.SetReleaseType(release.ReleaseTypeServer)
		} else {
			query.SetReleaseType(release.ReleaseTypeAgent)
		}

		if i%3 == 0 {
			query.SetOs("linux")
		} else {
			query.SetOs("windows")
		}

		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create releases")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *ReleasesTestSuite) TestGetLatestServerRelease() {
	r, err := suite.model.GetLatestServerRelease("stable")
	assert.NoError(suite.T(), err, "should get latest server release")
	assert.Equal(suite.T(), "0.1.6", r.Version, "should get 0.1.6 release")
}

func (suite *ReleasesTestSuite) TestGetLatestAgentRelease() {
	r, err := suite.model.GetLatestAgentRelease("stable")
	assert.NoError(suite.T(), err, "should get latest server release")
	assert.Equal(suite.T(), "0.1.5", r.Version, "should get 0.1.5 release")
}

func (suite *ReleasesTestSuite) TestGetServerReleases() {
	serverReleases := []string{"0.1.6", "0.1.4", "0.1.2", "0.1.0"}
	items, err := suite.model.GetServerReleases()
	assert.NoError(suite.T(), err, "should get server releases")
	assert.Equal(suite.T(), serverReleases, items, "should get 4 server releases")
}

func (suite *ReleasesTestSuite) TestGetAgentReleases() {
	agentReleases := []string{"0.1.5", "0.1.3", "0.1.1"}
	items, err := suite.model.GetAgentsReleases()
	assert.NoError(suite.T(), err, "should get server releases")
	assert.Equal(suite.T(), agentReleases, items, "should get 3 agent releases")
}

func (suite *ReleasesTestSuite) TestGetAgentsReleaseByType() {
	_, err := suite.model.GetAgentsReleaseByType(release.ReleaseTypeAgent, "stable", "linux", "amd64", "0.1.3")
	assert.NoError(suite.T(), err, "should get release agent linux 0.1.3")

	_, err = suite.model.GetAgentsReleaseByType(release.ReleaseTypeAgent, "stable", "windows", "amd64", "0.1.5")
	assert.NoError(suite.T(), err, "should get release agent windows 0.1.5")

	_, err = suite.model.GetAgentsReleaseByType(release.ReleaseTypeAgent, "stable", "linux", "amd64", "0.1.1")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should get error trying to get agent linux 0.1.1")
}

func (suite *ReleasesTestSuite) TestGetServersReleaseByType() {
	_, err := suite.model.GetServersReleaseByType(release.ReleaseTypeServer, "stable", "windows", "amd64", "0.1.2")
	assert.NoError(suite.T(), err, "should get release server windows 0.1.2")

	_, err = suite.model.GetServersReleaseByType(release.ReleaseTypeServer, "stable", "linux", "amd64", "0.1.6")
	assert.NoError(suite.T(), err, "should get release server linux 0.1.4")

	_, err = suite.model.GetServersReleaseByType(release.ReleaseTypeServer, "stable", "linux", "amd64", "0.1.4")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should get error trying to get server linux 0.1.6")
}

func (suite *ReleasesTestSuite) TestGetHigherAgentReleaseInstalled() {
	r, err := suite.model.GetHigherAgentReleaseInstalled()
	assert.NoError(suite.T(), err, "should get higher agent release installed")
	assert.Equal(suite.T(), "0.1.5", r.Version, "higher agent release should be 0.1.5")
}

func (suite *ReleasesTestSuite) TestCountOutdatedAgents() {
	count, err := suite.model.CountOutdatedAgents()
	assert.NoError(suite.T(), err, "should count outdated agents")
	assert.Equal(suite.T(), 2, count, "two agents should be outdated")
}

func (suite *ReleasesTestSuite) TestCountUpgradableAgents() {
	count, err := suite.model.CountUpgradableAgents("0.1.5")
	assert.NoError(suite.T(), err, "should count outdated agents")
	assert.Equal(suite.T(), 2, count, "two agents should be outdated")
}

func (suite *ReleasesTestSuite) TestSaveNewReleaseAvailable() {
	err := suite.model.SaveNewReleaseAvailable(release.ReleaseTypeAgent, openuem_nats.OpenUEMRelease{
		Version:         "0.2.0",
		Channel:         "stable",
		Summary:         "summary",
		ReleaseNotesURL: "url",
		ReleaseDate:     time.Now(),
		Files:           []openuem_nats.FileInfo{{Arch: "amd64", Os: "windows", FileURL: "url", Checksum: "checksum"}},
		IsCritical:      false,
	})
	assert.NoError(suite.T(), err, "should save new release")

	r, err := suite.model.GetLatestAgentRelease("stable")
	assert.NoError(suite.T(), err, "should get latest server release")
	assert.Equal(suite.T(), "0.2.0", r.Version, "should get 0.2.0 release")
}

func TestReleasesTestSuite(t *testing.T) {
	suite.Run(t, new(ReleasesTestSuite))
}
