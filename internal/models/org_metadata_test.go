package models

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-uem/ent/enttest"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OrgMetadataTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	p          partials.PaginationAndSort
	metadataId int
}

func (suite *OrgMetadataTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	for i := 0; i <= 6; i++ {
		m, err := client.OrgMetadata.Create().
			SetName(fmt.Sprintf("metadata%d", i)).
			SetDescription(fmt.Sprintf("metadata%d description", i)).
			Save(context.Background())
		assert.NoError(suite.T(), err, "set metadata description")
		suite.metadataId = m.ID
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *OrgMetadataTestSuite) TestGetAllOrgMetadata() {
	items, err := suite.model.GetAllOrgMetadata(&partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get all org metadata")

	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d", i), item.Name, fmt.Sprintf("name should be metadata%d", i))
	}
}

func (suite *OrgMetadataTestSuite) TestGetOrgMetadataByPage() {
	items, err := suite.model.GetOrgMetadataByPage(suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get org metadata by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d", i), item.Name)
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d description", i), item.Description)
	}

	suite.p.SortBy = "name"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetOrgMetadataByPage(suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get org metadata by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d", i), item.Name)
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d description", i), item.Description)
	}

	suite.p.SortBy = "name"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetOrgMetadataByPage(suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get org metadata by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d", 6-i), item.Name)
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d description", 6-i), item.Description)
	}

	suite.p.SortBy = "description"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetOrgMetadataByPage(suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get org metadata by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d", i), item.Name)
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d description", i), item.Description)
	}

	suite.p.SortBy = "description"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetOrgMetadataByPage(suite.p, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get org metadata by page")
	for i, item := range items {
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d", 6-i), item.Name)
		assert.Equal(suite.T(), fmt.Sprintf("metadata%d description", 6-i), item.Description)
	}
}

func (suite *OrgMetadataTestSuite) TestCountAllOrgMetadata() {
	count, err := suite.model.CountAllOrgMetadata(&partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all org metadata")
	assert.Equal(suite.T(), count, 7, "should have 7 metadata")
}

func (suite *OrgMetadataTestSuite) TestNewOrgMetadata() {
	err := suite.model.NewOrgMetadata("metadata7", "metadata7 description", &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should create metadata")

	items, err := suite.model.GetAllOrgMetadata(&partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get all org metadata")
	assert.Equal(suite.T(), "metadata7", items[7].Name, "should have metadata7 name")
	assert.Equal(suite.T(), "metadata7 description", items[7].Description, "should have metadata7 description")
}

func (suite *OrgMetadataTestSuite) TestUpdateOrgMetadata() {
	err := suite.model.UpdateOrgMetadata(suite.metadataId, "metadata7", "metadata7 description", &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should update metadata")

	items, err := suite.model.GetAllOrgMetadata(&partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should get all org metadata")
	assert.Equal(suite.T(), "metadata7", items[6].Name, "should have metadata7 name")
	assert.Equal(suite.T(), "metadata7 description", items[6].Description, "should have metadata7 description")
}

func (suite *OrgMetadataTestSuite) TestDeleteOrgMetadata() {
	err := suite.model.DeleteOrgMetadata(suite.metadataId, &partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should delete metadata")

	count, err := suite.model.CountAllOrgMetadata(&partials.CommonInfo{})
	assert.NoError(suite.T(), err, "should count all org metadata")
	assert.Equal(suite.T(), count, 6, "should have 6 metadata")
}

func TestOrgMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(OrgMetadataTestSuite))
}
