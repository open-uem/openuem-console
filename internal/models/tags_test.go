package models

import (
	"context"
	"fmt"
	"testing"

	"github.com/doncicuto/openuem-console/internal/views/partials"
	"github.com/doncicuto/openuem_ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TagsTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
	tagId int
}

func (suite *TagsTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	agent, err := client.Agent.Create().SetID("agent1").SetOs("windows").SetHostname("agent1").AddTagIDs().Save(context.Background())
	assert.NoError(suite.T(), err, "should create agent")

	for i := 0; i <= 6; i++ {
		tag, err := client.Tag.Create().SetTag(fmt.Sprintf("Tag%d", i)).SetDescription(fmt.Sprintf("My tag %d", i)).SetColor(fmt.Sprintf("#f%df%df%d", i, i, i)).Save(context.Background())
		assert.NoError(suite.T(), err)
		suite.tagId = tag.ID
		if i%2 == 0 {
			err := client.Agent.UpdateOneID(agent.ID).AddTagIDs(tag.ID).Exec(context.Background())
			assert.NoError(suite.T(), err, "should update agent to add tag")
		}
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *TagsTestSuite) TestGetAllTags() {
	tags, err := suite.model.GetAllTags()

	assert.NoError(suite.T(), err, "should get all tags")

	for i, tag := range tags {
		assert.Equal(suite.T(), fmt.Sprintf("Tag%d", i), tag.Tag, fmt.Sprintf("tag should be Tag%d", i))
		assert.Equal(suite.T(), fmt.Sprintf("My tag %d", i), tag.Description, fmt.Sprintf("tag description should be My tag %d", i))
		assert.Equal(suite.T(), fmt.Sprintf("#f%df%df%d", i, i, i), tag.Color, fmt.Sprintf("tag color should be #f%df%df%d", i, i, i))
	}
}

func (suite *TagsTestSuite) TestGetAppliedTags() {
	tags, err := suite.model.GetAppliedTags()

	assert.NoError(suite.T(), err, "should get all tags applied")

	assert.Equal(suite.T(), 4, len(tags), "four tags should be applied")
}

func (suite *TagsTestSuite) TestGetTagsByPage() {
	tags, err := suite.model.GetTagsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get all tags by page")
	assert.Equal(suite.T(), 5, len(tags), "five tags should be retrieved")

	suite.p.SortBy = "tag"
	suite.p.SortOrder = "desc"
	tags, err = suite.model.GetTagsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get tags by page")
	assert.Equal(suite.T(), "Tag6", tags[0].Tag, "tag should be Tag6")
	assert.Equal(suite.T(), "Tag5", tags[1].Tag, "tag should be Tag5")
	assert.Equal(suite.T(), "Tag4", tags[2].Tag, "tag should be Tag4")
	assert.Equal(suite.T(), "Tag3", tags[3].Tag, "tag should be Tag3")
	assert.Equal(suite.T(), "Tag2", tags[4].Tag, "tag should be Tag2")
	assert.Equal(suite.T(), "My tag 6", tags[0].Description, "description should be My tag 6")
	assert.Equal(suite.T(), "My tag 5", tags[1].Description, "description should be My tag 5")
	assert.Equal(suite.T(), "My tag 4", tags[2].Description, "description should be My tag 4")
	assert.Equal(suite.T(), "My tag 3", tags[3].Description, "description should be My tag 3")
	assert.Equal(suite.T(), "My tag 2", tags[4].Description, "description should be My tag 2")
	assert.Equal(suite.T(), "#f6f6f6", tags[0].Color, "color should be #f6f6f6")
	assert.Equal(suite.T(), "#f5f5f5", tags[1].Color, "color should be #f5f5f5")
	assert.Equal(suite.T(), "#f4f4f4", tags[2].Color, "color should be #f4f4f4")
	assert.Equal(suite.T(), "#f3f3f3", tags[3].Color, "color should be #f3f3f3")
	assert.Equal(suite.T(), "#f2f2f2", tags[4].Color, "color should be #f2f2f2")

	suite.p.SortBy = "tag"
	suite.p.SortOrder = "asc"
	tags, err = suite.model.GetTagsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get tags by page")
	assert.Equal(suite.T(), "Tag0", tags[0].Tag, "tag should be Tag0")
	assert.Equal(suite.T(), "Tag1", tags[1].Tag, "tag should be Tag1")
	assert.Equal(suite.T(), "Tag2", tags[2].Tag, "tag should be Tag2")
	assert.Equal(suite.T(), "Tag3", tags[3].Tag, "tag should be Tag3")
	assert.Equal(suite.T(), "Tag4", tags[4].Tag, "tag should be Tag4")
	assert.Equal(suite.T(), "My tag 0", tags[0].Description, "description should be My tag 0")
	assert.Equal(suite.T(), "My tag 1", tags[1].Description, "description should be My tag 1")
	assert.Equal(suite.T(), "My tag 2", tags[2].Description, "description should be My tag 2")
	assert.Equal(suite.T(), "My tag 3", tags[3].Description, "description should be My tag 3")
	assert.Equal(suite.T(), "My tag 4", tags[4].Description, "description should be My tag 4")
	assert.Equal(suite.T(), "#f0f0f0", tags[0].Color, "color should be #f0f0f0")
	assert.Equal(suite.T(), "#f1f1f1", tags[1].Color, "color should be #f1f1f1")
	assert.Equal(suite.T(), "#f2f2f2", tags[2].Color, "color should be #f2f2f2")
	assert.Equal(suite.T(), "#f3f3f3", tags[3].Color, "color should be #f3f3f3")
	assert.Equal(suite.T(), "#f4f4f4", tags[4].Color, "color should be #f4f4f4")

	suite.p.SortBy = "description"
	suite.p.SortOrder = "desc"
	tags, err = suite.model.GetTagsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get tags by page")
	assert.Equal(suite.T(), "Tag6", tags[0].Tag, "tag should be Tag6")
	assert.Equal(suite.T(), "Tag5", tags[1].Tag, "tag should be Tag5")
	assert.Equal(suite.T(), "Tag4", tags[2].Tag, "tag should be Tag4")
	assert.Equal(suite.T(), "Tag3", tags[3].Tag, "tag should be Tag3")
	assert.Equal(suite.T(), "Tag2", tags[4].Tag, "tag should be Tag2")
	assert.Equal(suite.T(), "My tag 6", tags[0].Description, "description should be My tag 6")
	assert.Equal(suite.T(), "My tag 5", tags[1].Description, "description should be My tag 5")
	assert.Equal(suite.T(), "My tag 4", tags[2].Description, "description should be My tag 4")
	assert.Equal(suite.T(), "My tag 3", tags[3].Description, "description should be My tag 3")
	assert.Equal(suite.T(), "My tag 2", tags[4].Description, "description should be My tag 2")
	assert.Equal(suite.T(), "#f6f6f6", tags[0].Color, "color should be #f6f6f6")
	assert.Equal(suite.T(), "#f5f5f5", tags[1].Color, "color should be #f5f5f5")
	assert.Equal(suite.T(), "#f4f4f4", tags[2].Color, "color should be #f4f4f4")
	assert.Equal(suite.T(), "#f3f3f3", tags[3].Color, "color should be #f3f3f3")
	assert.Equal(suite.T(), "#f2f2f2", tags[4].Color, "color should be #f2f2f2")

	suite.p.SortBy = "description"
	suite.p.SortOrder = "asc"
	suite.p.CurrentPage = 2
	tags, err = suite.model.GetTagsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get tags by page")
	assert.Equal(suite.T(), "Tag5", tags[0].Tag, "tag should be Tag5")
	assert.Equal(suite.T(), "Tag6", tags[1].Tag, "tag should be Tag6")
	assert.Equal(suite.T(), "My tag 5", tags[0].Description, "description should be My tag 5")
	assert.Equal(suite.T(), "My tag 6", tags[1].Description, "description should be My tag 6")
	assert.Equal(suite.T(), "#f5f5f5", tags[0].Color, "color should be #f5f5f5")
	assert.Equal(suite.T(), "#f6f6f6", tags[1].Color, "color should be #f6f6f6")
}

func (suite *TagsTestSuite) TestCountAllTags() {
	count, err := suite.model.CountAllTags()
	assert.NoError(suite.T(), err, "should count all tags")
	assert.Equal(suite.T(), 7, count, "tags count should be 7")
}

func (suite *TagsTestSuite) TestNewTag() {
	err := suite.model.NewTag("Tag8", "My tag 8", "#f8f8f8")
	assert.NoError(suite.T(), err, "should create a new tags")

	count, err := suite.model.CountAllTags()
	assert.NoError(suite.T(), err, "should count all tags")
	assert.Equal(suite.T(), 8, count, "tags count should be 8")

	tags, err := suite.model.GetAllTags()
	assert.NoError(suite.T(), err, "should get all tags")

	assert.Equal(suite.T(), "Tag8", tags[len(tags)-1].Tag, "tag should be Tag8")
	assert.Equal(suite.T(), "My tag 8", tags[len(tags)-1].Description, "tag description should be My tag 8")
	assert.Equal(suite.T(), "#f8f8f8", tags[len(tags)-1].Color, "tag color should be #f8f8f8")
}

func (suite *TagsTestSuite) TestUpdateTag() {
	err := suite.model.UpdateTag(suite.tagId, "Tag8", "My tag 8", "#f8f8f8")
	assert.NoError(suite.T(), err, "should update a tag")

	tags, err := suite.model.GetAllTags()
	assert.NoError(suite.T(), err, "should get all tags")

	assert.Equal(suite.T(), "Tag8", tags[len(tags)-1].Tag, "tag should be Tag8")
	assert.Equal(suite.T(), "My tag 8", tags[len(tags)-1].Description, "tag description should be My tag 8")
	assert.Equal(suite.T(), "#f8f8f8", tags[len(tags)-1].Color, "tag color should be #f8f8f8")
}

func (suite *TagsTestSuite) TestDeleteTag() {
	err := suite.model.DeleteTag(suite.tagId)
	assert.NoError(suite.T(), err, "should delete a tag")

	count, err := suite.model.CountAllTags()
	assert.NoError(suite.T(), err, "should count all tags")
	assert.Equal(suite.T(), 6, count, "tags count should be 6")
}

func TestTagsTestSuite(t *testing.T) {
	suite.Run(t, new(TagsTestSuite))
}
