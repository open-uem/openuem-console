package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ModelTestSuite struct {
	suite.Suite
}

func (suite *ModelTestSuite) TestNewModel() {
	_, err := New("file:ent?mode=memory&_fk=1", "sqlite3")
	assert.NoError(suite.T(), err, "should create model")

	_, err = New("postgres://localhost:1111/test", "pgx")
	assert.Error(suite.T(), err, "pgx should raise error")
}

func (suite *ModelTestSuite) TestCloseModel() {
	m, err := New("file:ent?mode=memory&_fk=1", "sqlite3")
	assert.NoError(suite.T(), err, "should create model")
	err = m.Close()
	assert.NoError(suite.T(), err, "should close model")
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(ModelTestSuite))
}
