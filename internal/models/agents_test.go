package models

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	openuem_ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AgentsTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	p          partials.PaginationAndSort
	tags       []int
	commonInfo *partials.CommonInfo
}

func (suite *AgentsTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	t, err := suite.model.CreateDefaultTenant()
	assert.NoError(suite.T(), err, "should create default tenant")

	s, err := suite.model.CreateDefaultSite(t)
	assert.NoError(suite.T(), err, "should create default site")

	suite.commonInfo = &partials.CommonInfo{TenantID: strconv.Itoa(t.ID), SiteID: strconv.Itoa(s.ID)}

	r, err := client.Release.Create().
		SetArch("amd64").
		SetChannel("stable").
		SetOs("windows").
		SetVersion("0.1.0").
		Save(context.Background())
	assert.NoError(suite.T(), err, "should create a release")

	for i := 0; i <= 6; i++ {
		query := client.Agent.Create().
			SetID(fmt.Sprintf("agent%d", i)).
			SetHostname(fmt.Sprintf("agent%d", i)).
			SetOs("windows").
			SetReleaseID(r.ID).
			SetNickname(fmt.Sprintf("agent%d", i)).
			SetLastContact(time.Now()).
			SetIP(fmt.Sprintf("192.168.1.%d", i)).
			SetUpdateTaskExecution(time.Now()).
			SetUpdateTaskDescription("update").
			SetUpdateTaskVersion("0.2.0").
			AddSiteIDs(s.ID)

		if i%2 == 0 {
			query.SetVnc("TigerVNC")
			query.SetAgentStatus(agent.AgentStatusEnabled)
			query.SetUpdateTaskStatus("Success")
			query.SetUpdateTaskResult("Success")
		} else {
			query.SetUpdateTaskStatus("Error")
			query.SetUpdateTaskResult("Error")
			if i == 1 {
				query.SetAgentStatus(agent.AgentStatusWaitingForAdmission)
			} else {
				query.SetAgentStatus(agent.AgentStatusDisabled)
			}
		}
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create agent")
	}

	for i := 0; i <= 6; i++ {
		tag, err := client.Tag.Create().SetTag(fmt.Sprintf("Tag%d", i)).SetDescription(fmt.Sprintf("My tag %d", i)).SetColor(fmt.Sprintf("#f%df%df%d", i, i, i)).Save(context.Background())
		assert.NoError(suite.T(), err)
		if i%2 == 0 {
			err := client.Agent.UpdateOneID(fmt.Sprintf("agent%d", i)).AddTagIDs(tag.ID).Exec(context.Background())
			assert.NoError(suite.T(), err, "should update agent to add tag")
		}
		suite.tags = append(suite.tags, tag.ID)
	}

	for i := 0; i <= 6; i++ {
		query := client.Antivirus.Create().SetName("antivirus")
		if i%2 == 0 {
			query.SetIsActive(true)
		} else {
			query.SetIsActive(false)
		}

		if i%3 == 0 {
			query.SetIsUpdated(false)
		} else {
			query.SetIsUpdated(true)
		}

		err := query.SetOwnerID(fmt.Sprintf("agent%d", i)).Exec(context.Background())
		assert.NoError(suite.T(), err, "should create antivirus")
	}

	for i := 0; i <= 6; i++ {
		query := client.SystemUpdate.Create().SetLastInstall(time.Now()).SetLastSearch(time.Now())
		if i%2 == 0 {
			query.SetPendingUpdates(true)
		} else {
			query.SetPendingUpdates(false)
		}

		if i%3 == 0 {
			query.SetSystemUpdateStatus(openuem_nats.NOTIFY_BEFORE_DOWNLOAD)
		} else {
			query.SetSystemUpdateStatus(openuem_nats.NOTIFY_BEFORE_INSTALLATION)
		}

		err := query.SetOwnerID(fmt.Sprintf("agent%d", i)).Exec(context.Background())
		assert.NoError(suite.T(), err, "should create system update")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *AgentsTestSuite) TestGetAllAgents() {
	items, err := suite.model.GetAllAgents(filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all agents")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID)
	}
}

func (suite *AgentsTestSuite) TestGetAgentsByPage() {
	excludeWaitingForAdmissionAgents := true
	items, err := suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID)
	}

	excludeWaitingForAdmissionAgents = false
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID)
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID)
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID)
	}

	suite.p.SortBy = "os"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID)
	}

	suite.p.SortBy = "os"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID)
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID)
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID)
	}

	suite.p.SortBy = "last_contact"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID)
	}

	suite.p.SortBy = "last_contact"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID)
	}

	suite.p.SortBy = "status"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	assert.Equal(suite.T(), "agent3", items[0].ID)
	assert.Equal(suite.T(), "agent5", items[1].ID)
	assert.Equal(suite.T(), "agent0", items[2].ID)
	assert.Equal(suite.T(), "agent2", items[3].ID)
	assert.Equal(suite.T(), "agent4", items[4].ID)

	suite.p.SortBy = "status"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	assert.Equal(suite.T(), "agent1", items[0].ID)
	assert.Equal(suite.T(), "agent0", items[1].ID)
	assert.Equal(suite.T(), "agent2", items[2].ID)
	assert.Equal(suite.T(), "agent4", items[3].ID)
	assert.Equal(suite.T(), "agent6", items[4].ID)

	suite.p.SortBy = "ip_address"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID)
	}

	suite.p.SortBy = "ip_address"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAgentsByPage(suite.p, filters.AgentFilter{Nickname: "agent"}, excludeWaitingForAdmissionAgents, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID)
	}
}

func (suite *AgentsTestSuite) TestGetAgentById() {
	var err error

	item, err := suite.model.GetAgentById("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agent by id")
	assert.Equal(suite.T(), "agent1", item.Nickname)

	_, err = suite.model.GetAgentById("agent7", suite.commonInfo)
	assert.Error(suite.T(), err, "should not get agent by id")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise is not found error")
}

func (suite *AgentsTestSuite) TestCountAllAgents() {
	count, err := suite.model.CountAllAgents(filters.AgentFilter{}, true, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 6, count, "should count 6 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{Nickname: "agent"}, true, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 6, count, "should count 6 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{Nickname: "agent"}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 7, count, "should count 7 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{AgentOSVersions: []string{"windows"}}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 7, count, "should count 7 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{ContactFrom: "2024-01-01", ContactTo: "2034-01-01"}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 7, count, "should count 7 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{AgentStatusOptions: []string{"Enabled"}}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 4, count, "should count 4 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{AgentStatusOptions: []string{"Disabled"}}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 2, count, "should count 2 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{AgentStatusOptions: []string{"WaitingForAdmission"}}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 1, count, "should count 1 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{AgentStatusOptions: []string{"WaitingForAdmission"}}, true, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 0, count, "should count 0 agents")

	count, err = suite.model.CountAllAgents(filters.AgentFilter{Tags: []int{suite.tags[0]}}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 1, count, "should count 1 agents")
}

func (suite *AgentsTestSuite) TestCountAgentsReportedLast24h() {
	count, err := suite.model.CountAgentsReportedLast24h(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count agents that reported in last 24h")
	assert.Equal(suite.T(), 6, count, "should count 6 agents that reported in last 24h")
}

func (suite *AgentsTestSuite) TestCountAgentsNotReportedLast24h() {
	count, err := suite.model.CountAgentsNotReportedLast24h(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count agents that not reported in last 24h")
	assert.Equal(suite.T(), 0, count, "should count 6 agents that not reported in last 24h")
}

func (suite *AgentsTestSuite) TestDeleteAgent() {
	err := suite.model.DeleteAgent("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should delete agent")

	err = suite.model.DeleteAgent("agent2", suite.commonInfo)
	assert.NoError(suite.T(), err, "should delete agent")

	count, err := suite.model.CountAllAgents(filters.AgentFilter{}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 5, count, "should count 5 agents")
}

func (suite *AgentsTestSuite) TestCountAgentsByOS() {
	items, err := suite.model.CountAgentsByOS(suite.commonInfo)
	assert.NoError(suite.T(), err, "should get os versions")
	assert.Equal(suite.T(), 1, len(items), "should get 1 os")
	assert.Equal(suite.T(), "windows", items[0].OS, "should get windows os")
	assert.Equal(suite.T(), 6, items[0].Count, "should get 6 agents")
}

func (suite *AgentsTestSuite) TestGetAgentsUsedOSes() {
	items, err := suite.model.GetAgentsUsedOSes(suite.commonInfo, filters.AgentFilter{})
	assert.NoError(suite.T(), err, "should get agents oses")
	assert.Equal(suite.T(), []string{"windows"}, items, "should get windows")
}

func (suite *AgentsTestSuite) TestEnableAgent() {
	err := suite.model.EnableAgent("agent3", suite.commonInfo)
	assert.NoError(suite.T(), err, "should enable agent")

	err = suite.model.EnableAgent("agent5", suite.commonInfo)
	assert.NoError(suite.T(), err, "should enable agent")

	count, err := suite.model.CountAllAgents(filters.AgentFilter{AgentStatusOptions: []string{"Enabled"}}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 6, count, "should count 6 agents")
}

func (suite *AgentsTestSuite) TestDisableAgent() {
	err := suite.model.DisableAgent("agent0", suite.commonInfo)
	assert.NoError(suite.T(), err, "should disable agent")

	err = suite.model.DisableAgent("agent2", suite.commonInfo)
	assert.NoError(suite.T(), err, "should disable agent")

	count, err := suite.model.CountAllAgents(filters.AgentFilter{AgentStatusOptions: []string{"Disabled"}}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 4, count, "should count 4 agents")
}

func (suite *AgentsTestSuite) TestCountDisabledAgents() {
	count, err := suite.model.CountDisabledAgents(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count disabled agents")
	assert.Equal(suite.T(), 2, count, "should count 3 disabled agents")
}

func (suite *AgentsTestSuite) TestAddTagToAgent() {
	err := suite.model.AddTagToAgent("agent0", strconv.Itoa(suite.tags[1]), &partials.CommonInfo{TenantID: "1", SiteID: "1"})
	assert.NoError(suite.T(), err, "should add tag to agent")
	count, err := suite.model.CountAllAgents(filters.AgentFilter{Tags: []int{suite.tags[0], suite.tags[1]}}, false, &partials.CommonInfo{TenantID: "1", SiteID: "1"})
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 1, count, "should count 1 agents")
}

func (suite *AgentsTestSuite) TestRemoveTagFromAgent() {
	err := suite.model.RemoveTagFromAgent("agent0", strconv.Itoa(suite.tags[0]), suite.commonInfo)
	assert.NoError(suite.T(), err, "should remove tag from agent")
	count, err := suite.model.CountAllAgents(filters.AgentFilter{Tags: []int{suite.tags[0]}}, false, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all agents")
	assert.Equal(suite.T(), 0, count, "should count 0 agents")
}

func (suite *AgentsTestSuite) TestCountDisabledAntivirusAgents() {
	count, err := suite.model.CountDisabledAntivirusAgents(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count disabled antivirus")
	assert.Equal(suite.T(), 2, count, "should count 2 disabled antivirus")
}

func (suite *AgentsTestSuite) TestCountOutdatedAntivirusDatabaseAgents() {
	count, err := suite.model.CountOutdatedAntivirusDatabaseAgents(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count outdated antivirus")
	assert.Equal(suite.T(), 3, count, "should count 3 outdated antivirus")
}

func (suite *AgentsTestSuite) TestCountVNCSupportedAgents() {
	count, err := suite.model.CountVNCSupportedAgents(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count VNC supported agents")
	assert.Equal(suite.T(), 4, count, "should count 4 agents with supported VNC")
}

func (suite *AgentsTestSuite) TestCountWaitingForAdmissionAgents() {
	count, err := suite.model.CountWaitingForAdmissionAgents(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count waiting for admission agents")
	assert.Equal(suite.T(), 1, count, "should count 1 agent waiting for admission")
}

func (suite *AgentsTestSuite) TestAgentsExists() {
	exists, err := suite.model.AgentsExists(suite.commonInfo)
	assert.NoError(suite.T(), err, "should check if agents exists")
	assert.Equal(suite.T(), true, exists, "should check if agents exists")
}

func (suite *AgentsTestSuite) TestDeleteAllAgents() {
	count, err := suite.model.DeleteAllAgents(suite.commonInfo)
	assert.NoError(suite.T(), err, "should delete all agents")
	assert.Equal(suite.T(), 7, count, "should delete 7 agents")

	exists, err := suite.model.AgentsExists(suite.commonInfo)
	assert.NoError(suite.T(), err, "should check if agents exists")
	assert.Equal(suite.T(), false, exists, "agents should not exist")
}

func (suite *AgentsTestSuite) TestCountPendingUpdateAgents() {
	count, err := suite.model.CountPendingUpdateAgents(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count pending update agents")
	assert.Equal(suite.T(), 4, count, "should count 4 agents with pending updates")
}

func (suite *AgentsTestSuite) TestCountNoAutoupdateAgents() {
	count, err := suite.model.CountNoAutoupdateAgents(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count no autoupdate agents")
	assert.Equal(suite.T(), 6, count, "should count 7 agents with no system auto update")
}

func (suite *AgentsTestSuite) TestCountAllUpdateAgents() {
	count, err := suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 6, count, "should count 6 agents")

	count, err = suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{Nickname: "agent0"}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 1, count, "should count 1 agents")

	count, err = suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{Releases: []string{"0.1.0"}}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 6, count, "should count 6 agents")

	count, err = suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{Tags: []int{suite.tags[0]}}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 1, count, "should count 1 agents")

	count, err = suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{TaskStatus: []string{"Success", "Error"}}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 6, count, "should count 6 agents")

	count, err = suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{TaskStatus: []string{"Error"}}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 2, count, "should count 2 agents")

	count, err = suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{TaskResult: "Error"}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 2, count, "should count 2 agents")

	count, err = suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{TaskLastExecutionFrom: "2024-01-01", TaskLastExecutionTo: "2034-01-01"}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 6, count, "should count 6 agents")
}

func (suite *AgentsTestSuite) TestGetAllUpdateAgents() {
	items, err := suite.model.GetAllUpdateAgents(filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get all update agents")
	for i, item := range items {
		if i < 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
		}
		if i > 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i+1), item.Nickname)
		}
	}
}

func (suite *AgentsTestSuite) TestSaveAgentUpdateInfo() {
	err := suite.model.SaveAgentUpdateInfo("agent3", "Success", "description", "0.2.0", suite.commonInfo)
	assert.NoError(suite.T(), err, "should save agent update info")

	count, err := suite.model.CountAllUpdateAgents(filters.UpdateAgentsFilter{TaskResult: "Error"}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all update agents")
	assert.Equal(suite.T(), 1, count, "should count 1 agents")
}

func (suite *AgentsTestSuite) TestGetUpdateAgentsByPage() {
	items, err := suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Nickname)
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		if i < 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
		}
		if i > 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i+1), item.Nickname)
		}
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Nickname)
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		if i < 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
		}
		if i > 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i+1), item.Nickname)
		}
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		if i < 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
		}
		if i > 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i+1), item.Nickname)
		}
	}

	suite.p.SortBy = "taskStatus"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	assert.Equal(suite.T(), "agent3", items[0].Nickname)
	assert.Equal(suite.T(), "agent5", items[1].Nickname)
	assert.Equal(suite.T(), "agent0", items[2].Nickname)
	assert.Equal(suite.T(), "agent2", items[3].Nickname)
	assert.Equal(suite.T(), "agent4", items[4].Nickname)

	suite.p.SortBy = "taskStatus"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	assert.Equal(suite.T(), "agent0", items[0].Nickname)
	assert.Equal(suite.T(), "agent2", items[1].Nickname)
	assert.Equal(suite.T(), "agent4", items[2].Nickname)
	assert.Equal(suite.T(), "agent6", items[3].Nickname)
	assert.Equal(suite.T(), "agent3", items[4].Nickname)

	suite.p.SortBy = "taskDescription"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		if i < 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
		}
		if i > 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i+1), item.Nickname)
		}
	}

	suite.p.SortBy = "taskDescription"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		if i < 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
		}
		if i > 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i+1), item.Nickname)
		}
	}

	suite.p.SortBy = "taskLastExecution"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		if i < 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Nickname)
		}
		if i > 1 {
			assert.Equal(suite.T(), fmt.Sprintf("agent%d", i+1), item.Nickname)
		}
	}

	suite.p.SortBy = "taskLastExecution"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Nickname)
	}

	suite.p.SortBy = "taskResult"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	assert.Equal(suite.T(), "agent3", items[0].Nickname)
	assert.Equal(suite.T(), "agent5", items[1].Nickname)
	assert.Equal(suite.T(), "agent0", items[2].Nickname)
	assert.Equal(suite.T(), "agent2", items[3].Nickname)
	assert.Equal(suite.T(), "agent4", items[4].Nickname)

	suite.p.SortBy = "taskResult"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetUpdateAgentsByPage(suite.p, filters.UpdateAgentsFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get agents by page")
	assert.Equal(suite.T(), "agent0", items[0].Nickname)
	assert.Equal(suite.T(), "agent2", items[1].Nickname)
	assert.Equal(suite.T(), "agent4", items[2].Nickname)
	assert.Equal(suite.T(), "agent6", items[3].Nickname)
	assert.Equal(suite.T(), "agent3", items[4].Nickname)

}

func TestAgentsTestSuite(t *testing.T) {
	suite.Run(t, new(AgentsTestSuite))
}
