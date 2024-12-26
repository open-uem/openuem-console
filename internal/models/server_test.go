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

type ServersTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *ServersTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	for i := 0; i <= 6; i++ {
		err := client.User.Create().SetID(fmt.Sprintf("user%d", i)).SetName(fmt.Sprintf("User%d", i)).Exec(context.Background())
		assert.NoError(suite.T(), err)
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func TestServersTestSuite(t *testing.T) {
	suite.Run(t, new(ServersTestSuite))
}
