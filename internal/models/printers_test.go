package models

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PrintersTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	commonInfo *partials.CommonInfo
}

func (suite *PrintersTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	t, err := suite.model.CreateDefaultTenant()
	assert.NoError(suite.T(), err, "should create default tenant")

	s, err := suite.model.CreateDefaultSite(t)
	assert.NoError(suite.T(), err, "should create default site")

	suite.commonInfo = &partials.CommonInfo{TenantID: strconv.Itoa(t.ID), SiteID: strconv.Itoa(s.ID)}

	err = client.Agent.Create().SetID("agent1").SetHostname("agent1").SetOs("windows").SetNickname("agent1").AddSiteIDs(s.ID).Exec(context.Background())
	assert.NoError(suite.T(), err, "should create agent")

	for i := 0; i <= 6; i++ {
		err := client.Printer.Create().
			SetName(fmt.Sprintf("printer%d", i)).
			SetOwnerID("agent1").
			Exec(context.Background())
		assert.NoError(suite.T(), err)
	}
}

func (suite *PrintersTestSuite) TestCountDifferentPrinters() {
	count, err := suite.model.CountDifferentPrinters(suite.commonInfo)
	assert.NoError(suite.T(), err, "should count different printers")
	assert.Equal(suite.T(), 7, count, "should count 7 different printers")
}

func TestPrintersTestSuite(t *testing.T) {
	suite.Run(t, new(PrintersTestSuite))
}
