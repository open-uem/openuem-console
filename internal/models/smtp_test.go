package models

import (
	"context"
	"testing"

	"github.com/open-uem/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SMTPTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	settingsId int
}

func (suite *SMTPTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	settings, err := suite.model.Client.Settings.Create().Save(context.Background())
	assert.NoError(suite.T(), err, "should create initial settings")
	suite.settingsId = settings.ID
}

func (suite *SMTPTestSuite) TestGetSMTPSettings() {
	settings, err := suite.model.GetSMTPSettings()
	assert.NoError(suite.T(), err, "should get default SMTP settings")

	assert.Equal(suite.T(), "", settings.SMTPServer, "server should be empty")
	assert.Equal(suite.T(), 587, settings.SMTPPort, "port should be 587")
	assert.Equal(suite.T(), "", settings.SMTPUser, "user should be empty")
	assert.Equal(suite.T(), "", settings.SMTPPassword, "password should be empty")
	assert.Equal(suite.T(), "LOGIN", settings.SMTPAuth, "auth should be LOGIN")
	assert.Equal(suite.T(), "", settings.MessageFrom, "message from should be empty")
}

func (suite *SMTPTestSuite) TestUpdateSMTPSettings() {
	newSettings := SMTPSettings{
		ID:       suite.settingsId,
		Server:   "smtp.example.com",
		Auth:     "PLAIN",
		Port:     465,
		User:     "test",
		Password: "test",
		MailFrom: "test@example.com",
	}

	err := suite.model.UpdateSMTPSettings(&newSettings)
	assert.NoError(suite.T(), err, "should update SMTP settings")

	settings, err := suite.model.GetSMTPSettings()
	assert.NoError(suite.T(), err, "should get updated SMTP settings")

	assert.Equal(suite.T(), "smtp.example.com", settings.SMTPServer, "server should be smtp.example.com")
	assert.Equal(suite.T(), 465, settings.SMTPPort, "port should be 465")
	assert.Equal(suite.T(), "test", settings.SMTPUser, "user should be test")
	assert.Equal(suite.T(), "test", settings.SMTPPassword, "password should be test")
	assert.Equal(suite.T(), "LOGIN", settings.SMTPAuth, "auth should be PLAIN")
	assert.Equal(suite.T(), "test@example.com", settings.MessageFrom, "message from should be test@example.com")
}

func TestSMTPTestSuite(t *testing.T) {
	suite.Run(t, new(SMTPTestSuite))
}
