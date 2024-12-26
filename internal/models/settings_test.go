package models

import (
	"context"
	"testing"

	"github.com/doncicuto/openuem_ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SettingsTestSuite struct {
	suite.Suite
	t          enttest.TestingT
	model      Model
	settingsId int
}

func (suite *SettingsTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	settings, err := client.Settings.Create().Save(context.Background())
	assert.NoError(suite.T(), err)

	suite.settingsId = settings.ID
}

func (suite *SettingsTestSuite) TestGetMaxUploadSize() {
	maxUploadSize, err := suite.model.GetMaxUploadSize()
	assert.NoError(suite.T(), err, "should get max upload size")
	assert.Equal(suite.T(), "512M", maxUploadSize, "default max upload size should be 512M")
}

func (suite *SettingsTestSuite) TestUpdateMaxUploadSizeSetting() {
	err := suite.model.UpdateMaxUploadSizeSetting(suite.settingsId, "128M")
	assert.NoError(suite.T(), err, "should update max upload size setting")

	setting, err := suite.model.GetMaxUploadSize()
	assert.NoError(suite.T(), err, "should get max upload size setting")

	assert.Equal(suite.T(), "128M", setting, "max upload size setting should be 128M")
}

func (suite *SettingsTestSuite) TestGetDefaultCountry() {
	defaultCountry, err := suite.model.GetDefaultCountry()
	assert.NoError(suite.T(), err, "should get default country")
	assert.Equal(suite.T(), "ES", defaultCountry, "default country should be ES")
}

func (suite *SettingsTestSuite) TestUpdateCountrySetting() {
	err := suite.model.UpdateCountrySetting(suite.settingsId, "FR")
	assert.NoError(suite.T(), err, "should update country setting")

	setting, err := suite.model.GetDefaultCountry()
	assert.NoError(suite.T(), err, "should get country setting")

	assert.Equal(suite.T(), "FR", setting, "country setting should be FR")
}

func (suite *SettingsTestSuite) TestGetDefaultAgentFrequency() {
	defaultAgentFrequency, err := suite.model.GetDefaultAgentFrequency()
	assert.NoError(suite.T(), err, "should get default agent frequencue")
	assert.Equal(suite.T(), 60, defaultAgentFrequency, "default agency frequency should be 60")
}

func (suite *SettingsTestSuite) TestUpdateAgentFrequency() {
	err := suite.model.UpdateAgentFrequency(suite.settingsId, 30)
	assert.NoError(suite.T(), err, "should update agent frequency setting")

	setting, err := suite.model.GetDefaultAgentFrequency()
	assert.NoError(suite.T(), err, "should get agent frequency setting")

	assert.Equal(suite.T(), 30, setting, "agent frequency should be 30")
}

func (suite *SettingsTestSuite) TestGetDefaultRefreshTime() {
	defaultRefreshTime, err := suite.model.GetDefaultRefreshTime()
	assert.NoError(suite.T(), err, "should get default refresh time")
	assert.Equal(suite.T(), 5, defaultRefreshTime, "default refresh time should be 5")
}

func (suite *SettingsTestSuite) TestUpdateRefreshTimeSetting() {
	err := suite.model.UpdateRefreshTimeSetting(suite.settingsId, 10)
	assert.NoError(suite.T(), err, "should update console refresh time setting")

	setting, err := suite.model.GetDefaultRefreshTime()
	assert.NoError(suite.T(), err, "should get console refresh time setting")

	assert.Equal(suite.T(), 10, setting, "console refresh time should be 10")
}

func (suite *SettingsTestSuite) TestGetDefaultUpdateChannel() {
	defaultUpdateChannel, err := suite.model.GetDefaultUpdateChannel()
	assert.NoError(suite.T(), err, "should get default update channel")
	assert.Equal(suite.T(), "stable", defaultUpdateChannel, "default update channel should be stable")
}

func (suite *SettingsTestSuite) TestUpdateOpenUEMChannel() {
	err := suite.model.UpdateOpenUEMChannel(suite.settingsId, "devel")
	assert.NoError(suite.T(), err, "should update channel setting")

	setting, err := suite.model.GetDefaultUpdateChannel()
	assert.NoError(suite.T(), err, "should get update channel setting")

	assert.Equal(suite.T(), "devel", setting, "update channel should be devel")
}

func (suite *SettingsTestSuite) TestGetDefaultSessionLifetime() {
	defaultSessionLifetime, err := suite.model.GetDefaultSessionLifetime()
	assert.NoError(suite.T(), err, "should get default session lifetime")
	assert.Equal(suite.T(), 1440, defaultSessionLifetime, "default session lifetime should be 1440")
}

func (suite *SettingsTestSuite) TestUpdateSessionLifetime() {
	err := suite.model.UpdateSessionLifetime(suite.settingsId, 180)
	assert.NoError(suite.T(), err, "should update session lifetime setting")

	setting, err := suite.model.GetDefaultSessionLifetime()
	assert.NoError(suite.T(), err, "should get session lifetime setting")

	assert.Equal(suite.T(), 180, setting, "session lifetime should be 180")
}

func (suite *SettingsTestSuite) TestGetDefaultUserCertDuration() {
	defaultUserCertDuration, err := suite.model.GetDefaultUserCertDuration()
	assert.NoError(suite.T(), err, "should get default user cert duration")
	assert.Equal(suite.T(), 1, defaultUserCertDuration, "default user cert duration should be 1")
}

func (suite *SettingsTestSuite) TestUpdateUserCertDurationSetting() {
	err := suite.model.UpdateUserCertDurationSetting(suite.settingsId, 2)
	assert.NoError(suite.T(), err, "should update user cert duration setting")

	setting, err := suite.model.GetDefaultUserCertDuration()
	assert.NoError(suite.T(), err, "should get user cert duration setting")

	assert.Equal(suite.T(), 2, setting, "user cert duration should be 2")
}

func (suite *SettingsTestSuite) TestGetNATSTimeout() {
	natsTimeout, err := suite.model.GetNATSTimeout()
	assert.NoError(suite.T(), err, "should get nats timeout")
	assert.Equal(suite.T(), 20, natsTimeout, "default nats timeout should be 20")
}

func (suite *SettingsTestSuite) TestDefaultRequestVNCPIN() {
	requestVNCPin, err := suite.model.GetDefaultRequestVNCPIN()
	assert.NoError(suite.T(), err, "should get default request vnc pin setting")
	assert.Equal(suite.T(), true, requestVNCPin, "by default request vnc pin should be true")
}

func (suite *SettingsTestSuite) TestGetGeneralSettings() {

	err := suite.model.Client.Settings.Update().
		SetMaxUploadSize("128M").
		SetCountry("FR").
		SetAgentReportFrequenceInMinutes(30).
		SetRefreshTimeInMinutes(10).
		SetUpdateChannel("devel").
		SetSessionLifetimeInMinutes(440).
		SetUserCertYearsValid(2).
		SetNatsRequestTimeoutSeconds(60).
		SetRequestVncPin(false).
		Exec(context.Background())
	assert.NoError(suite.T(), err, "settings should be updated")

	settings, err := suite.model.GetGeneralSettings()
	assert.NoError(suite.T(), err, "settings should be retrieved")

	assert.Equal(suite.T(), "128M", settings.MaxUploadSize, "default max upload size should be 128M")
	assert.Equal(suite.T(), "FR", settings.Country, "country should be FR")
	assert.Equal(suite.T(), 30, settings.AgentReportFrequenceInMinutes, "agent report frequency should be 60 minutes")
	assert.Equal(suite.T(), 10, settings.RefreshTimeInMinutes, "refresh time should be 10 minutes")
	assert.Equal(suite.T(), "devel", settings.UpdateChannel, "update channel should be devel")
	assert.Equal(suite.T(), 440, settings.SessionLifetimeInMinutes, "session lifetime in minutes should be 440")
	assert.Equal(suite.T(), 2, settings.UserCertYearsValid, "user cert years should be 2")
	assert.Equal(suite.T(), 60, settings.NatsRequestTimeoutSeconds, "nats timeout should be 60")
	assert.Equal(suite.T(), false, settings.RequestVncPin, "request vnc pin should be false")
}

func (suite *SettingsTestSuite) TestCreateInitialSettings() {
	_, err := suite.model.Client.Settings.Delete().Exec(context.Background())
	assert.NoError(suite.T(), err, "settings should be deleted")

	err = suite.model.CreateInitialSettings()
	assert.NoError(suite.T(), err, "initial settings should be deleted")

	settings, err := suite.model.GetGeneralSettings()
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "512M", settings.MaxUploadSize, "default max upload size should be 512M")
	assert.Equal(suite.T(), "ES", settings.Country, "default country should be ES")
	assert.Equal(suite.T(), 60, settings.AgentReportFrequenceInMinutes, "default agent report frequency should be 60 minutes")
	assert.Equal(suite.T(), 5, settings.RefreshTimeInMinutes, "default refresh time should be 5 minutes")
	assert.Equal(suite.T(), "stable", settings.UpdateChannel, "default update channel should be stable")
	assert.Equal(suite.T(), 1440, settings.SessionLifetimeInMinutes, "default session lifetime in minutes should be 1440")
	assert.Equal(suite.T(), 1, settings.UserCertYearsValid, "default user cert years should be 2")
	assert.Equal(suite.T(), 20, settings.NatsRequestTimeoutSeconds, "default nats timeout should be 20")
	assert.Equal(suite.T(), true, settings.RequestVncPin, "default request vnc pin should be true")
}

func TestSettingsTestSuite(t *testing.T) {
	suite.Run(t, new(SettingsTestSuite))
}
