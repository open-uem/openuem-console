package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	openuem_ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/enttest"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
	t        enttest.TestingT
	model    Model
	p        partials.PaginationAndSort
	tenantID int
}

func (suite *UserTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	// Create a tenant for testing
	tenant, err := client.Tenant.Create().
		SetDescription("TestTenant").
		SetIsDefault(true).
		Save(context.Background())
	assert.NoError(suite.T(), err)
	suite.tenantID = tenant.ID

	for i := 0; i <= 6; i++ {
		err := client.User.Create().
			SetID(fmt.Sprintf("user%d", i)).
			SetName(fmt.Sprintf("User %d", i)).
			SetEmail(fmt.Sprintf("user%d@example.com", i)).
			SetPhone(fmt.Sprintf("%d%d%d %d%d%d %d%d%d", i, i, i, i, i, i, i, i, i)).
			SetCountry(fmt.Sprintf("E%d", i)).
			SetRegister(fmt.Sprintf("Register%d", i)).
			SetCreated(time.Now()).
			Exec(context.Background())
		assert.NoError(suite.T(), err)

		// Assign user to tenant
		err = client.UserTenant.Create().
			SetUserID(fmt.Sprintf("user%d", i)).
			SetTenantID(tenant.ID).
			SetRole("user").
			SetIsDefault(true).
			Exec(context.Background())
		assert.NoError(suite.T(), err)
	}

	suite.p = partials.PaginationAndSort{CurrentPage: 1, PageSize: 5}
}

func (suite *UserTestSuite) TestCountAllUsers() {
	f := filters.UserFilter{}
	count, err := suite.model.CountAllUsers(f, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 7, count, "should count 7 users")

	f = filters.UserFilter{Username: "user6"}
	count, err = suite.model.CountAllUsers(f, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 1, count, "should count 1 users")

	f = filters.UserFilter{Name: "User"}
	count, err = suite.model.CountAllUsers(f, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 7, count, "should count 7 users")

	f = filters.UserFilter{Email: "user1@example.com"}
	count, err = suite.model.CountAllUsers(f, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 1, count, "should count 1 users")

	f = filters.UserFilter{Phone: "111 111 111"}
	count, err = suite.model.CountAllUsers(f, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 1, count, "should count 1 users")

	f = filters.UserFilter{CreatedFrom: "2024-01-01", CreatedTo: "2030-01-01"}
	count, err = suite.model.CountAllUsers(f, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 7, count, "should count 7 users")

	f = filters.UserFilter{ModifiedFrom: "2024-01-01", ModifiedTo: "2030-01-01"}
	count, err = suite.model.CountAllUsers(f, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 7, count, "should count 7 users")

	f = filters.UserFilter{RegisterOptions: []string{"Register0", "Register1", "Register6"}}
	count, err = suite.model.CountAllUsers(f, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 3, count, "should count 3 users")
}

func (suite *UserTestSuite) TestGetUsersByPage() {
	users, err := suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), 5, len(users), "five users should be retrieved")

	suite.p.SortBy = "uid"
	suite.p.SortOrder = "asc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user0", users[0].ID, "user id should be user0")
	assert.Equal(suite.T(), "user1", users[1].ID, "user id should be user1")
	assert.Equal(suite.T(), "user2", users[2].ID, "user id should be user2")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user4", users[4].ID, "user id should be user4")

	suite.p.SortBy = "uid"
	suite.p.SortOrder = "desc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user6", users[0].ID, "user id should be user6")
	assert.Equal(suite.T(), "user5", users[1].ID, "user id should be user5")
	assert.Equal(suite.T(), "user4", users[2].ID, "user id should be user4")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user2", users[4].ID, "user id should be user2")

	suite.p.SortBy = "name"
	suite.p.SortOrder = "asc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "User 0", users[0].Name, "user name should be User 0")
	assert.Equal(suite.T(), "User 1", users[1].Name, "user name should be User 1")
	assert.Equal(suite.T(), "User 2", users[2].Name, "user name should be User 2")
	assert.Equal(suite.T(), "User 3", users[3].Name, "user name should be User 3")
	assert.Equal(suite.T(), "User 4", users[4].Name, "user name should be User 4")

	suite.p.SortBy = "name"
	suite.p.SortOrder = "desc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "User 6", users[0].Name, "user name should be User 6")
	assert.Equal(suite.T(), "User 5", users[1].Name, "user name should be User 5")
	assert.Equal(suite.T(), "User 4", users[2].Name, "user name should be User 4")
	assert.Equal(suite.T(), "User 3", users[3].Name, "user name should be User 3")
	assert.Equal(suite.T(), "User 2", users[4].Name, "user name should be User 2")

	suite.p.SortBy = "email"
	suite.p.SortOrder = "asc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user0@example.com", users[0].Email, "user email should be user0@example.com")
	assert.Equal(suite.T(), "user1@example.com", users[1].Email, "user email should be user1@example.com")
	assert.Equal(suite.T(), "user2@example.com", users[2].Email, "user email should be user2@example.com")
	assert.Equal(suite.T(), "user3@example.com", users[3].Email, "user email should be user3@example.com")
	assert.Equal(suite.T(), "user4@example.com", users[4].Email, "user email should be user4@example.com")

	suite.p.SortBy = "email"
	suite.p.SortOrder = "desc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user6@example.com", users[0].Email, "user email should be user6@example.com")
	assert.Equal(suite.T(), "user5@example.com", users[1].Email, "user email should be user5@example.com")
	assert.Equal(suite.T(), "user4@example.com", users[2].Email, "user email should be user4@example.com")
	assert.Equal(suite.T(), "user3@example.com", users[3].Email, "user email should be user3@example.com")
	assert.Equal(suite.T(), "user2@example.com", users[4].Email, "user email should be user2@example.com")

	suite.p.SortBy = "phone"
	suite.p.SortOrder = "asc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "000 000 000", users[0].Phone, "user phone should be 000 000 000")
	assert.Equal(suite.T(), "111 111 111", users[1].Phone, "user phone should be 111 111 111")
	assert.Equal(suite.T(), "222 222 222", users[2].Phone, "user phone should be 222 222 222")
	assert.Equal(suite.T(), "333 333 333", users[3].Phone, "user phone should be 333 333 333")
	assert.Equal(suite.T(), "444 444 444", users[4].Phone, "user phone should be 444 444 444")

	suite.p.SortBy = "phone"
	suite.p.SortOrder = "desc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "666 666 666", users[0].Phone, "user phone should be 666 666 666")
	assert.Equal(suite.T(), "555 555 555", users[1].Phone, "user phone should be 555 555 555")
	assert.Equal(suite.T(), "444 444 444", users[2].Phone, "user phone should be 444 444 444")
	assert.Equal(suite.T(), "333 333 333", users[3].Phone, "user phone should be 333 333 333")
	assert.Equal(suite.T(), "222 222 222", users[4].Phone, "user phone should be 222 222 222")

	suite.p.SortBy = "country"
	suite.p.SortOrder = "asc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user0", users[0].ID, "user id should be user0")
	assert.Equal(suite.T(), "user1", users[1].ID, "user id should be user1")
	assert.Equal(suite.T(), "user2", users[2].ID, "user id should be user2")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user4", users[4].ID, "user id should be user4")

	suite.p.SortBy = "country"
	suite.p.SortOrder = "desc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user6", users[0].ID, "user id should be user6")
	assert.Equal(suite.T(), "user5", users[1].ID, "user id should be user5")
	assert.Equal(suite.T(), "user4", users[2].ID, "user id should be user4")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user2", users[4].ID, "user id should be user2")

	suite.p.SortBy = "register"
	suite.p.SortOrder = "asc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user0", users[0].ID, "user id should be user0")
	assert.Equal(suite.T(), "user1", users[1].ID, "user id should be user1")
	assert.Equal(suite.T(), "user2", users[2].ID, "user id should be user2")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user4", users[4].ID, "user id should be user4")

	suite.p.SortBy = "register"
	suite.p.SortOrder = "desc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user6", users[0].ID, "user id should be user6")
	assert.Equal(suite.T(), "user5", users[1].ID, "user id should be user5")
	assert.Equal(suite.T(), "user4", users[2].ID, "user id should be user4")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user2", users[4].ID, "user id should be user2")

	suite.p.SortBy = "created"
	suite.p.SortOrder = "asc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user0", users[0].ID, "user id should be user0")
	assert.Equal(suite.T(), "user1", users[1].ID, "user id should be user1")
	assert.Equal(suite.T(), "user2", users[2].ID, "user id should be user2")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user4", users[4].ID, "user id should be user4")

	suite.p.SortBy = "created"
	suite.p.SortOrder = "desc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user6", users[0].ID, "user id should be user6")
	assert.Equal(suite.T(), "user5", users[1].ID, "user id should be user5")
	assert.Equal(suite.T(), "user4", users[2].ID, "user id should be user4")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user2", users[4].ID, "user id should be user2")

	suite.p.SortBy = "modified"
	suite.p.SortOrder = "asc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user0", users[0].ID, "user id should be user0")
	assert.Equal(suite.T(), "user1", users[1].ID, "user id should be user1")
	assert.Equal(suite.T(), "user2", users[2].ID, "user id should be user2")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user4", users[4].ID, "user id should be user4")

	suite.p.SortBy = "modified"
	suite.p.SortOrder = "desc"
	users, err = suite.model.GetUsersByPage(suite.p, filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should get all users by page")
	assert.Equal(suite.T(), "user6", users[0].ID, "user id should be user6")
	assert.Equal(suite.T(), "user5", users[1].ID, "user id should be user5")
	assert.Equal(suite.T(), "user4", users[2].ID, "user id should be user4")
	assert.Equal(suite.T(), "user3", users[3].ID, "user id should be user3")
	assert.Equal(suite.T(), "user2", users[4].ID, "user id should be user2")
}

func (suite *UserTestSuite) TestUserExists() {
	exists, err := suite.model.UserExists("user1")
	assert.NoError(suite.T(), err, "should check if users exists")
	assert.Equal(suite.T(), true, exists, "user should exist")

	exists, err = suite.model.UserExists("user8")
	assert.NoError(suite.T(), err, "should check if users exists")
	assert.Equal(suite.T(), false, exists, "user should not exist")
}

func (suite *UserTestSuite) TestEmailExists() {
	exists, err := suite.model.EmailExists("user2@example.com")
	assert.NoError(suite.T(), err, "should check if email exists")
	assert.Equal(suite.T(), true, exists, "email should exist")

	exists, err = suite.model.EmailExists("user8@example.com")
	assert.NoError(suite.T(), err, "should check if email exists")
	assert.Equal(suite.T(), false, exists, "email should not exist")
}

func (suite *UserTestSuite) TestAddLocalUser() {
	_, err := suite.model.AddUser("user7", "User7", "user7@example.com", "", "ES", "certificate")
	assert.NoError(suite.T(), err, "should add a new user")

	user, err := suite.model.GetUserById("user7")
	assert.NoError(suite.T(), err, "should get recently created user")
	assert.Equal(suite.T(), "user7", user.ID, "user should have user7 id")
	assert.Equal(suite.T(), "User7", user.Name, "user should have User7 name")
	assert.Equal(suite.T(), "user7@example.com", user.Email, "user should have user7@example.com email")
	assert.Equal(suite.T(), "", user.Phone, "user should have empty phone")
	assert.Equal(suite.T(), "ES", user.Country, "user should have ES country")
	assert.Equal(suite.T(), "users.pending_email_confirmation", user.Register, "user should have users.pending_email_confirmation register status")
}

func (suite *UserTestSuite) TestAddOIDCUser() {
	_, err := suite.model.AddUser("user7", "User7", "user7@example.com", "", "ES", "certificate")
	assert.NoError(suite.T(), err, "should add a new user")

	user, err := suite.model.GetUserById("user7")
	assert.NoError(suite.T(), err, "should get recently created user")
	assert.Equal(suite.T(), "user7", user.ID, "user should have user7 id")
	assert.Equal(suite.T(), "User7", user.Name, "user should have User7 name")
	assert.Equal(suite.T(), "user7@example.com", user.Email, "user should have user7@example.com email")
	assert.Equal(suite.T(), "", user.Phone, "user should have empty phone")
	assert.Equal(suite.T(), "ES", user.Country, "user should have ES country")
	assert.Equal(suite.T(), "users.review_request", user.Register, "user should have users.review_request register status")
}

func (suite *UserTestSuite) TestAddImportedUser() {
	err := suite.model.AddImportedUser("user7", "User7", "user7@example.com", "", "ES", false)
	assert.NoError(suite.T(), err, "should add a new user")

	user, err := suite.model.GetUserById("user7")
	assert.NoError(suite.T(), err, "should get recently created user")
	assert.Equal(suite.T(), "user7", user.ID, "user should have user7 id")
	assert.Equal(suite.T(), "User7", user.Name, "user should have User7 name")
	assert.Equal(suite.T(), "user7@example.com", user.Email, "user should have user7@example.com email")
	assert.Equal(suite.T(), "", user.Phone, "user should have empty phone")
	assert.Equal(suite.T(), "ES", user.Country, "user should have ES country")
	assert.Equal(suite.T(), "users.certificate_sent", user.Register, "user should have users.certificate_sent register status")
}

func (suite *UserTestSuite) TestAddImporteOIDCdUser() {
	err := suite.model.AddImportedUser("user7", "User7", "user7@example.com", "", "ES", true)
	assert.NoError(suite.T(), err, "should add a new user")

	user, err := suite.model.GetUserById("user7")
	assert.NoError(suite.T(), err, "should get recently created user")
	assert.Equal(suite.T(), "user7", user.ID, "user should have user7 id")
	assert.Equal(suite.T(), "User7", user.Name, "user should have User7 name")
	assert.Equal(suite.T(), "user7@example.com", user.Email, "user should have user7@example.com email")
	assert.Equal(suite.T(), "", user.Phone, "user should have empty phone")
	assert.Equal(suite.T(), "ES", user.Country, "user should have ES country")
	assert.Equal(suite.T(), "users.review_request", user.Register, "user should have users.review_request register status")
}

func (suite *UserTestSuite) TestUpdateUser() {
	err := suite.model.UpdateUser("user9", "User7", "user7@example.com", "", "ES")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "cannot update non existing user")

	err = suite.model.UpdateUser("user2", "User7", "user7@example.com", "", "ES")
	assert.NoError(suite.T(), err, "should update existing user")

	user, err := suite.model.GetUserById("user2")
	assert.NoError(suite.T(), err, "should get recently updated user")
	assert.Equal(suite.T(), "user2", user.ID, "user should have user2 id")
	assert.Equal(suite.T(), "User7", user.Name, "user should have User7 name")
	assert.Equal(suite.T(), "user7@example.com", user.Email, "user should have user7@example.com email")
	assert.Equal(suite.T(), "", user.Phone, "user should have empty phone")
	assert.Equal(suite.T(), "ES", user.Country, "user should have ES country")
}

func (suite *UserTestSuite) TestRegisterUser() {
	err := suite.model.RegisterUser("user7", "User7", "user7@example.com", "", "ES", "apassword", "certificate")
	assert.NoError(suite.T(), err, "should register a user")

	user, err := suite.model.GetUserById("user7")
	assert.NoError(suite.T(), err, "should get recently created user")
	assert.Equal(suite.T(), "user7", user.ID, "user should have user7 id")
	assert.Equal(suite.T(), "User7", user.Name, "user should have User7 name")
	assert.Equal(suite.T(), "user7@example.com", user.Email, "user should have user7@example.com email")
	assert.Equal(suite.T(), "", user.Phone, "user should have empty phone")
	assert.Equal(suite.T(), "ES", user.Country, "user should have ES country")
	assert.Equal(suite.T(), "apassword", user.CertClearPassword, "user should have apassword cert clear password")
}

func (suite *UserTestSuite) TestGetUserById() {
	_, err := suite.model.GetUserById("user9")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "cannot find non existing user")

	user, err := suite.model.GetUserById("user3")
	assert.NoError(suite.T(), err, "should get recently created user")
	assert.Equal(suite.T(), "user3", user.ID, "user should have user3 id")
	assert.Equal(suite.T(), "User 3", user.Name, "user should have User3 name")
	assert.Equal(suite.T(), "user3@example.com", user.Email, "user should have user3@example.com email")
}

func (suite *UserTestSuite) TestConfirmEmail() {
	err := suite.model.ConfirmEmail("user4")
	assert.NoError(suite.T(), err, "should confirm email")

	user, err := suite.model.GetUserById("user4")
	assert.NoError(suite.T(), err, "should get confirmed email user")
	assert.Equal(suite.T(), "user4", user.ID, "user should have user4 id")
	assert.Equal(suite.T(), "User 4", user.Name, "user should have User4 name")
	assert.Equal(suite.T(), true, user.EmailVerified, "user should have email verified")
	assert.Equal(suite.T(), openuem_nats.REGISTER_IN_REVIEW, user.Register)
}

func (suite *UserTestSuite) TestUserSetRevokedCertificate() {
	err := suite.model.UserSetRevokedCertificate("user4")
	assert.NoError(suite.T(), err, "should set register to certificate revoked")

	user, err := suite.model.GetUserById("user4")
	assert.NoError(suite.T(), err, "should get confirmed email user")
	assert.Equal(suite.T(), "user4", user.ID, "user should have user4 id")
	assert.Equal(suite.T(), "User 4", user.Name, "user should have User4 name")
	assert.Equal(suite.T(), openuem_nats.REGISTER_REVOKED, user.Register)
}

func (suite *UserTestSuite) TestConfirmLogIn() {
	err := suite.model.ConfirmLogIn("user5")
	assert.NoError(suite.T(), err, "should confirm user log in")

	user, err := suite.model.GetUserById("user5")
	assert.NoError(suite.T(), err, "should get confirmed log in user")
	assert.Equal(suite.T(), "user5", user.ID, "user should have user5 id")
	assert.Equal(suite.T(), "User 5", user.Name, "user should have User 5 name")
	assert.Equal(suite.T(), "", user.CertClearPassword, "user should have empty cert clear password")
	assert.Equal(suite.T(), openuem_nats.REGISTER_COMPLETE, user.Register)
}

func (suite *UserTestSuite) TestDeleteUser() {
	err := suite.model.DeleteUser("user6")
	assert.NoError(suite.T(), err, "should delete user")

	count, err := suite.model.CountAllUsers(filters.UserFilter{}, suite.tenantID)
	assert.NoError(suite.T(), err, "should count all users")
	assert.Equal(suite.T(), 6, count, "should count 6 users")
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
