package models

import (
	"context"

	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/settings"
)

func (m *Model) GetMaxUploadSize() (string, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldMaxUploadSize).Only(context.Background())
	if err != nil {
		return "", err
	}

	return settings.MaxUploadSize, nil
}

func (m *Model) UpdateMaxUploadSizeSetting(settingsId int, size string) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetMaxUploadSize(size).Exec(context.Background())
}

func (m *Model) GetNATSTimeout() (int, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldNatsRequestTimeoutSeconds).Only(context.Background())
	if err != nil {
		return 0, err
	}

	return settings.NatsRequestTimeoutSeconds, nil
}

func (m *Model) UpdateNATSTimeoutSetting(settingsId, timeout int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetNatsRequestTimeoutSeconds(timeout).Exec(context.Background())
}

func (m *Model) GetDefaultCountry() (string, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldCountry).Only(context.Background())
	if err != nil {
		return "", err
	}

	return settings.Country, nil
}

func (m *Model) UpdateCountrySetting(settingsId int, country string) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetCountry(country).Exec(context.Background())
}

func (m *Model) GetDefaultUserCertDuration() (int, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldUserCertYearsValid).Only(context.Background())
	if err != nil {
		return 0, err
	}

	return settings.UserCertYearsValid, nil
}

func (m *Model) UpdateUserCertDurationSetting(settingsId, years int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetUserCertYearsValid(years).Exec(context.Background())
}

func (m *Model) GetDefaultRefreshTime() (int, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldRefreshTimeInMinutes).Only(context.Background())
	if err != nil {
		return 0, err
	}

	return settings.RefreshTimeInMinutes, nil
}

func (m *Model) UpdateRefreshTimeSetting(settingsId, refresh int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetRefreshTimeInMinutes(refresh).Exec(context.Background())
}

func (m *Model) GetDefaultSessionLifetime() (int, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldSessionLifetimeInMinutes).Only(context.Background())
	if err != nil {
		return 0, err
	}

	return settings.SessionLifetimeInMinutes, nil
}

func (m *Model) UpdateSessionLifetime(settingsId, sessionLifetime int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetSessionLifetimeInMinutes(sessionLifetime).Exec(context.Background())
}

func (m *Model) GetDefaultAgentFrequency() (int, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldAgentReportFrequenceInMinutes).Only(context.Background())
	if err != nil {
		return 0, err
	}

	return settings.AgentReportFrequenceInMinutes, nil
}

func (m *Model) UpdateAgentFrequency(settingsId, frequency int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetAgentReportFrequenceInMinutes(frequency).Exec(context.Background())
}

func (m *Model) GetDefaultUpdateChannel() (string, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldUpdateChannel).Only(context.Background())
	if err != nil {
		return "", err
	}

	return settings.UpdateChannel, nil
}

func (m *Model) UpdateRequestVNCPIN(settingsId int, requestPIN bool) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetRequestVncPin(requestPIN).Exec(context.Background())
}

func (m *Model) GetDefaultRequestVNCPIN() (bool, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldRequestVncPin).Only(context.Background())
	if err != nil {
		return true, err
	}

	return settings.RequestVncPin, nil
}

func (m *Model) UpdateOpenUEMChannel(settingsId int, updateChannel string) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetUpdateChannel(updateChannel).Exec(context.Background())
}

func (m *Model) GetGeneralSettings() (*openuem_ent.Settings, error) {

	query := m.Client.Settings.Query().Select(
		settings.FieldID,
		settings.FieldCountry,
		settings.FieldMaxUploadSize,
		settings.FieldUserCertYearsValid,
		settings.FieldNatsRequestTimeoutSeconds,
		settings.FieldRefreshTimeInMinutes,
		settings.FieldSessionLifetimeInMinutes,
		settings.FieldUpdateChannel,
		settings.FieldAgentReportFrequenceInMinutes,
		settings.FieldRequestVncPin,
	)

	settings, err := query.Only(context.Background())

	if err != nil {
		if !openuem_ent.IsNotFound(err) {
			return nil, err
		} else {
			if err := m.Client.Settings.Create().Exec(context.Background()); err != nil {
				return nil, err
			}
			return query.Only(context.Background())
		}
	}

	return settings, nil
}

func (m *Model) CreateInitialSettings() error {
	nSettings, err := m.Client.Settings.Query().Count(context.Background())
	if err != nil {
		return err
	}

	if nSettings == 0 {
		return m.Client.Settings.Create().Exec(context.Background())
	}
	return nil
}

type GeneralSettings struct {
	ID              int
	Country         string
	MaxUploadSize   string
	UserCertYears   int
	NATSTimeout     int
	Refresh         int
	SessionLifetime int
	UpdateChannel   string
	AgentFrequency  int
	RequestVNCPIN   bool
}
