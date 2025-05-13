package models

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AntivirusTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *AntivirusTestSuite) SetupTest() {
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
		query := client.Antivirus.Create().
			SetName(fmt.Sprintf("antivirus%d", i)).
			SetIsActive(true).
			SetOwnerID(fmt.Sprintf("agent%d", i))

		if i%2 == 0 {
			query.SetIsActive(true)
			query.SetIsUpdated(true)
		} else {
			query.SetIsActive(true)
			query.SetIsUpdated(false)
		}

		err := query.Exec(context.Background())
		assert.NoError(suite.T(), err, "should create antivirus")
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *AntivirusTestSuite) TestCountAllAntiviri() {
	count, err := suite.model.CountAllAntiviri(filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all antiviri")
	assert.Equal(suite.T(), 7, count, "should count 7 antiviri")

	f := filters.AntivirusFilter{Hostname: "agent5"}
	count, err = suite.model.CountAllAntiviri(f, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all antiviri")
	assert.Equal(suite.T(), 1, count, "should count 1 antiviri")

	f = filters.AntivirusFilter{AgentOSVersions: []string{"windows"}}
	count, err = suite.model.CountAllAntiviri(f, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all antiviri")
	assert.Equal(suite.T(), 7, count, "should count 7 antiviri")

	f = filters.AntivirusFilter{AntivirusNameOptions: []string{"antivirus1", "antivirus2", "antivirus3"}}
	count, err = suite.model.CountAllAntiviri(f, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all antiviri")
	assert.Equal(suite.T(), 3, count, "should count 3 antiviri")

	f = filters.AntivirusFilter{AntivirusEnabledOptions: []string{"Enabled"}}
	count, err = suite.model.CountAllAntiviri(f, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all antiviri")
	assert.Equal(suite.T(), 7, count, "should count 7 antiviri")

	f = filters.AntivirusFilter{AntivirusEnabledOptions: []string{"Disabled"}}
	count, err = suite.model.CountAllAntiviri(f, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all antiviri")
	assert.Equal(suite.T(), 0, count, "should count 0 antiviri")

	f = filters.AntivirusFilter{AntivirusUpdatedOptions: []string{"UpdatedYes"}}
	count, err = suite.model.CountAllAntiviri(f, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all antiviri")
	assert.Equal(suite.T(), 4, count, "should count 4 antiviri")

	f = filters.AntivirusFilter{AntivirusUpdatedOptions: []string{"UpdatedNo"}}
	count, err = suite.model.CountAllAntiviri(f, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all antiviri")
	assert.Equal(suite.T(), 3, count, "should count 3 antiviri")
}

func (suite *AntivirusTestSuite) TestGetAntiviriByPage() {
	items, err := suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Hostname)
	}

	suite.p.SortBy = "hostname"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "hostname"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Hostname)
	}

	suite.p.SortBy = "agentOS"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "agentOS"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "antivirusName"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "antivirusName"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6-i), item.Hostname)
	}

	suite.p.SortBy = "antivirusEnabled"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "antivirusEnabled"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("agent%d", i), item.Hostname)
	}

	suite.p.SortBy = "antivirusUpdated"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[0].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 3), items[1].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 5), items[2].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[3].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[4].Hostname)

	suite.p.SortBy = "antivirusUpdated"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetAntiviriByPage(suite.p, filters.AntivirusFilter{}, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get antiviri by page")
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 0), items[0].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 2), items[1].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 4), items[2].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 6), items[3].Hostname)
	assert.Equal(suite.T(), fmt.Sprintf("agent%d", 1), items[4].Hostname)
}

func (suite *AntivirusTestSuite) TestGetDetectedAntiviri() {
	antiviri := []string{"antivirus0", "antivirus1", "antivirus2", "antivirus3", "antivirus4", "antivirus5", "antivirus6"}
	av, err := suite.model.GetDetectedAntiviri(&partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should detect antiviri")
	assert.Equal(suite.T(), antiviri, av, "should get detected antiviri")
}

func TestAntivirusTestSuite(t *testing.T) {
	suite.Run(t, new(AntivirusTestSuite))
}
