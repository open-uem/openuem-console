package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem_ent"
	"github.com/open-uem/openuem_ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SessionsTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *SessionsTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	for i := 0; i <= 6; i++ {
		err := client.User.Create().SetID(fmt.Sprintf("user%d", i)).SetName(fmt.Sprintf("User%d", i)).Exec(context.Background())
		assert.NoError(suite.T(), err)
	}

	for i := 0; i <= 6; i++ {
		err := client.Sessions.Create().SetData([]byte(fmt.Sprintf("session%d", i))).SetExpiry(time.Now()).SetID(fmt.Sprintf("token%d", i)).SetOwnerID(fmt.Sprintf("user%d", i)).Exec(context.Background())
		assert.NoError(suite.T(), err)
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *SessionsTestSuite) TestCountAllSessions() {
	nSessions, err := suite.model.CountAllSessions()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 7, nSessions, "number of sessions should be 7")
}

func (suite *SessionsTestSuite) TestGetSessionsByPage() {
	sessions, err := suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get sessions by page")
	assert.Equal(suite.T(), 5, len(sessions), "number of sessions should be 5")

	suite.p.SortBy = "token"
	suite.p.SortOrder = "desc"
	sessions, err = suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get sessions by page")
	assert.Equal(suite.T(), "token6", sessions[0].ID, "first session should have token6")
	assert.Equal(suite.T(), "token2", sessions[4].ID, "last session should have token2")

	suite.p.SortBy = "token"
	suite.p.SortOrder = "asc"
	sessions, err = suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get sessions by page")
	assert.Equal(suite.T(), "token0", sessions[0].ID, "first session should have token0")
	assert.Equal(suite.T(), "token4", sessions[4].ID, "last session should have token4")

	suite.p.SortBy = "expiry"
	suite.p.SortOrder = "desc"
	sessions, err = suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get sessions by page")
	assert.Equal(suite.T(), "token6", sessions[0].ID, "first session should have token6")
	assert.Equal(suite.T(), "token2", sessions[suite.p.PageSize-1].ID, "last session should have token2")

	suite.p.SortBy = "expiry"
	suite.p.SortOrder = "asc"
	sessions, err = suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get sessions by page")
	assert.Equal(suite.T(), "token0", sessions[0].ID, "first session should have token0")
	assert.Equal(suite.T(), "token4", sessions[suite.p.PageSize-1].ID, "last session should have token4")

	suite.p.SortBy = "uid"
	suite.p.SortOrder = "desc"
	sessions, err = suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get sessions by page")
	assert.Equal(suite.T(), "token6", sessions[0].ID, "first session should have token6")
	assert.Equal(suite.T(), "token2", sessions[suite.p.PageSize-1].ID, "last session should have token2")

	suite.p.SortBy = "uid"
	suite.p.SortOrder = "asc"
	sessions, err = suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get sessions by page")
	assert.Equal(suite.T(), "token0", sessions[0].ID, "first session should have token0")
	assert.Equal(suite.T(), "token4", sessions[suite.p.PageSize-1].ID, "last session should have token4")

	suite.p.PageSize = 10
	sessions, err = suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err, "should get sessions by page")
	assert.Equal(suite.T(), 7, len(sessions), "number of sessions should be 7")
}

func (suite *SessionsTestSuite) TestDeleteSession() {
	err := suite.model.DeleteSession("token1")
	assert.NoError(suite.T(), err, "session with token token1 should be deleted")

	err = suite.model.DeleteSession("token2")
	assert.NoError(suite.T(), err, "session with token token2 should be deleted")

	err = suite.model.DeleteSession("token3")
	assert.NoError(suite.T(), err, "session with token token3 should be deleted")

	err = suite.model.DeleteSession("token1")
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), openuem_ent.IsNotFound(err), true, "session with token token1 was deleted previously")
	}

	sessions, err := suite.model.GetSessionsByPage(suite.p)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), 4, len(sessions), "number of sessions should be 4")
}

func TestSessionsTestSuite(t *testing.T) {
	suite.Run(t, new(SessionsTestSuite))
}
