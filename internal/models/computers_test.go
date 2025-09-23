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
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ComputersTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	p          partials.PaginationAndSort
	tags       []int
	commonInfo *partials.CommonInfo
}

func (suite *ComputersTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	t, err := suite.model.CreateDefaultTenant()
	assert.NoError(suite.T(), err, "should create default tenant")

	s, err := suite.model.CreateDefaultSite(t)
	assert.NoError(suite.T(), err, "should create default site")

	suite.commonInfo = &partials.CommonInfo{TenantID: strconv.Itoa(t.ID), SiteID: strconv.Itoa(s.ID)}

	tags := []int{}
	for i := range 2 {
		tag, err := client.Tag.Create().SetTag(fmt.Sprintf("Tag%d", i)).SetDescription(fmt.Sprintf("My tag %d", i)).SetColor(fmt.Sprintf("#f%df%df%d", i, i, i)).Save(context.Background())
		assert.NoError(suite.T(), err, "should create tag")
		tags = append(tags, tag.ID)
	}
	suite.tags = tags

	for i := range 7 {
		query := client.Agent.Create().
			SetID(fmt.Sprintf("agent%d", i)).
			SetHostname(fmt.Sprintf("agent%d", i)).
			SetOs("windows").
			SetNickname(fmt.Sprintf("agent%d", i)).
			SetAgentStatus(agent.AgentStatusEnabled).
			AddSiteIDs(s.ID).
			SetLastContact(time.Now())
		if i%2 == 0 {
			query.AddTagIDs(tags[0])
		} else {
			query.AddTagIDs(tags[1])
		}
		_, err := query.Save(context.Background())
		assert.NoError(suite.T(), err, "should create agent")
	}

	for i := range 7 {
		err := client.OperatingSystem.Create().
			SetType("windows").
			SetUsername(fmt.Sprintf("user%d", i)).
			SetVersion(fmt.Sprintf("windows%d", i)).
			SetDescription(fmt.Sprintf("description%d", i)).
			SetOwnerID(fmt.Sprintf("agent%d", i)).
			Exec(context.Background())
		assert.NoError(suite.T(), err, "should create operating system")
	}

	for i := range 7 {
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

	for i := range 7 {
		query := client.NetworkAdapter.Create().
			SetName(fmt.Sprintf("network%d", i)).
			SetMACAddress(fmt.Sprintf("FF:FF:FF:FF:FF:F%d", i)).
			SetAddresses(fmt.Sprintf("192.168.1.%d", i)).
			SetSpeed("100Mbps").
			SetOwnerID("agent1")
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create network address")
	}

	for i := range 7 {
		query := client.Printer.Create().
			SetName(fmt.Sprintf("printer%d", i)).
			SetOwnerID("agent1")
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create printer")
	}

	for i := range 7 {
		query := client.LogicalDisk.Create().
			SetLabel(fmt.Sprintf("logicalDisk%d", i)).
			SetOwnerID("agent1")
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create logical disk")
	}

	for i := range 7 {
		query := client.PhysicalDisk.Create().
			SetDeviceID(fmt.Sprintf("physicalDisk%d", i)).
			SetModel(fmt.Sprintf("physicalDisk%d", i)).
			SetOwnerID("agent1")
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create physical disk")
	}

	for i := range 7 {
		query := client.Share.Create().
			SetName(fmt.Sprintf("share%d", i)).
			SetDescription(fmt.Sprintf("description%d", i)).
			SetOwnerID("agent1")
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create share")
	}

	for i := range 7 {
		query := client.Monitor.Create().
			SetManufacturer(fmt.Sprintf("manufacturer%d", i)).
			SetModel(fmt.Sprintf("model%d", i)).
			SetOwnerID("agent1")
		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create monitor")
	}

	for i := range 7 {
		err := client.App.Create().
			SetName("App").
			SetVersion("0.1.0").
			SetOwnerID(fmt.Sprintf("agent%d", i)).Exec(context.Background())
		assert.NoError(suite.T(), err, "should create app")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *ComputersTestSuite) TestGetComputersByPage() {
	items, err := suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID, fmt.Sprintf("agent ID should be %d", 6-i))
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID, fmt.Sprintf("agent ID should be %d", i))
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID, fmt.Sprintf("agent ID should be %d", 6-i))
	}

	suite.p.SortBy = "os"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID, fmt.Sprintf("agent ID should be %d", i))
	}

	suite.p.SortBy = "os"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID, fmt.Sprintf("agent ID should be %d", i))
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID, fmt.Sprintf("agent ID should be %d", i))
	}

	suite.p.SortBy = "version"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID, fmt.Sprintf("agent ID should be %d", 6-i))
	}

	suite.p.SortBy = "username"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID, fmt.Sprintf("agent ID should be %d", i))
	}

	suite.p.SortBy = "username"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID, fmt.Sprintf("agent ID should be %d", 6-i))
	}

	suite.p.SortBy = "manufacturer"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID, fmt.Sprintf("agent ID should be %d", i))
	}

	suite.p.SortBy = "manufacturer"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID, fmt.Sprintf("agent ID should be %d", 6-i))
	}

	suite.p.SortBy = "model"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.ID, fmt.Sprintf("agent ID should be %d", i))
	}

	suite.p.SortBy = "model"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.ID, fmt.Sprintf("agent ID should be %d", 6-i))
	}

	suite.p.SortBy = "nickname"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetComputersByPage(suite.p, filters.AgentFilter{Tags: []int{suite.tags[1]}}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computers by page")
	assert.Equal(suite.T(), 3, len(items), "should get three computers")
	assert.Equal(suite.T(), "agent1", items[0].ID, "agent shoud be agent1")
	assert.Equal(suite.T(), "agent3", items[1].ID, "agent shoud be agent3")
	assert.Equal(suite.T(), "agent5", items[2].ID, "agent shoud be agent5")
}

func (suite *ComputersTestSuite) TestCountAllComputers() {
	count, err := suite.model.CountAllComputers(filters.AgentFilter{}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 7, count, "should count 7 computers")

	f := filters.AgentFilter{Nickname: "agent"}
	count, err = suite.model.CountAllComputers(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 7, count, "should count 7 computers")

	f = filters.AgentFilter{AgentOSVersions: []string{"windows"}}
	count, err = suite.model.CountAllComputers(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 7, count, "should count 7 computers")

	f = filters.AgentFilter{ComputerManufacturers: []string{"manufacturer0", "manufacturer1", "manufacturer3"}}
	count, err = suite.model.CountAllComputers(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 3, count, "should count 3 computers")

	f = filters.AgentFilter{ComputerModels: []string{"model0", "model1", "model2", "model3"}}
	count, err = suite.model.CountAllComputers(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 4, count, "should count 4 computers")

	f = filters.AgentFilter{OSVersions: []string{"windows1", "windows4"}}
	count, err = suite.model.CountAllComputers(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 2, count, "should count 2 computers")

	f = filters.AgentFilter{Username: "user1"}
	count, err = suite.model.CountAllComputers(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 1, count, "should count 1 computers")

	f = filters.AgentFilter{WithApplication: "App"}
	count, err = suite.model.CountAllComputers(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 7, count, "should count 7 computers")

	f = filters.AgentFilter{Tags: []int{suite.tags[1]}}
	count, err = suite.model.CountAllComputers(f, suite.commonInfo)
	assert.NoError(suite.T(), err, "should count all computers")
	assert.Equal(suite.T(), 3, count, "should count 3 computers")
}

func (suite *ComputersTestSuite) TestGetAgentComputerInfo() {
	var err error

	item, err := suite.model.GetAgentComputerInfo("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should found agent1")
	assert.Equal(suite.T(), "manufacturer1", item.Edges.Computer.Manufacturer, "manufacturer should be manufacturer1")

	_, err = suite.model.GetAgentComputerInfo("agent7", suite.commonInfo)
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestGetAgentOSInfo() {
	var err error

	item, err := suite.model.GetAgentOSInfo("agent3", suite.commonInfo)
	assert.NoError(suite.T(), err, "should found agent3")
	assert.Equal(suite.T(), "windows3", item.Edges.Operatingsystem.Version, "version should be windows3")
	assert.Equal(suite.T(), "user3", item.Edges.Operatingsystem.Username, "user should be user3")

	_, err = suite.model.GetAgentOSInfo("agent7", suite.commonInfo)
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestGetAgentNetworkAdaptersInfo() {
	var err error

	item, err := suite.model.GetAgentNetworkAdaptersInfo("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should found agent1")
	assert.Equal(suite.T(), "network0", item.Edges.Networkadapters[0].Name, "network should be network0")
	assert.Equal(suite.T(), "FF:FF:FF:FF:FF:F0", item.Edges.Networkadapters[0].MACAddress, "version should be FF:FF:FF:FF:FF:F1")

	_, err = suite.model.GetAgentNetworkAdaptersInfo("agent7", suite.commonInfo)
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestGetAgentLogicalDisksInfo() {
	var err error

	item, err := suite.model.GetAgentLogicalDisksInfo("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should found agent1")
	assert.Equal(suite.T(), "logicalDisk0", item.Edges.Logicaldisks[0].Label, "logicalDisk should be logicalDisk0")
	assert.Equal(suite.T(), "logicalDisk1", item.Edges.Logicaldisks[1].Label, "logicalDisk should be logicalDisk1")

	_, err = suite.model.GetAgentLogicalDisksInfo("agent7", suite.commonInfo)
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestGetAgentPhysicalDisksInfo() {
	var err error

	item, err := suite.model.GetAgentPhysicalDisksInfo("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should found agent1")
	assert.Equal(suite.T(), "physicalDisk0", item.Edges.Physicaldisks[0].Model, "physicalDisk should be physicalDisk0")
	assert.Equal(suite.T(), "physicalDisk1", item.Edges.Physicaldisks[1].Model, "physicalDisk should be physicalDisk1")

	_, err = suite.model.GetAgentPhysicalDisksInfo("agent7", suite.commonInfo)
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestGetAgentSharesInfo() {
	var err error

	item, err := suite.model.GetAgentSharesInfo("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should found agent1")
	assert.Equal(suite.T(), "share2", item.Edges.Shares[2].Name, "share name should be share2")
	assert.Equal(suite.T(), "share3", item.Edges.Shares[3].Name, "share name should be share3")

	_, err = suite.model.GetAgentSharesInfo("agent7", suite.commonInfo)
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestGetAgentMonitorsInfo() {
	var err error

	item, err := suite.model.GetAgentMonitorsInfo("agent1", suite.commonInfo)
	assert.NoError(suite.T(), err, "should found agent1")
	assert.Equal(suite.T(), "manufacturer2", item.Edges.Monitors[2].Manufacturer, "manufacturer should be manufacturer2")
	assert.Equal(suite.T(), "model3", item.Edges.Monitors[3].Model, "model should be model3")

	_, err = suite.model.GetAgentMonitorsInfo("agent7", suite.commonInfo)
	assert.Error(suite.T(), err, "should not found agent7")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *ComputersTestSuite) TestSaveNotes() {
	err := suite.model.SaveNotes("agent1", "notes", suite.commonInfo)
	assert.NoError(suite.T(), err, "should save notes")

	items, err := suite.model.Client.Agent.Query().Where(agent.ID("agent1")).All(context.Background())
	assert.NoError(suite.T(), err, "should get computers by page")
	assert.Equal(suite.T(), 1, len(items), "should get one item")
	assert.Equal(suite.T(), "notes", items[0].Notes, "should get notes")
}

func (suite *ComputersTestSuite) TestGetComputerManufacturers() {
	allManufacturers := []string{"manufacturer0", "manufacturer1", "manufacturer2", "manufacturer3", "manufacturer4", "manufacturer5", "manufacturer6"}
	items, err := suite.model.GetComputerManufacturers(suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computer manufacturers")
	assert.Equal(suite.T(), 7, len(allManufacturers), "should get 7 manufacturers")
	assert.Equal(suite.T(), allManufacturers, items, "should get 7 manufacturers")
}

func (suite *ComputersTestSuite) TestGetComputerModels() {
	allModels := []string{"model1", "model2"}
	items, err := suite.model.GetComputerModels(filters.AgentFilter{ComputerManufacturers: []string{"manufacturer1", "manufacturer2"}}, suite.commonInfo)
	assert.NoError(suite.T(), err, "should get computer models")
	assert.Equal(suite.T(), allModels, items, "should get two computer models")
}

func (suite *ComputersTestSuite) TestCountDifferentVendor() {
	count, err := suite.model.CountDifferentVendor(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count different vendors")
	assert.Equal(suite.T(), 7, count, "should count 7 different vendors")
}

func TestComputersTestSuite(t *testing.T) {
	suite.Run(t, new(ComputersTestSuite))
}
