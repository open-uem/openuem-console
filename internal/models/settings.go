package models

import (
	"context"
	"strconv"

	openuem_ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/settings"
	"github.com/open-uem/ent/tenant"
)

type GeneralSettings struct {
	ID                       int
	Country                  string
	MaxUploadSize            string
	UserCertYears            int
	NATSTimeout              int
	Refresh                  int
	SessionLifetime          int
	UpdateChannel            string
	AgentFrequency           int
	RequestVNCPIN            bool
	Tag                      int
	WinGetFrequency          int
	UseWinget                bool
	UseFlatpak               bool
	SFTPDisabled             bool
	RemoteAssistanceDisabled bool
	DetectRemoteAgents       bool
	AutoAdmitAgents          bool
}

func (m *Model) GetMaxUploadSize() (string, error) {
	var err error

	settings, err := m.Client.Settings.Query().Select(settings.FieldMaxUploadSize).Where(settings.Not(settings.HasTenant())).Only(context.Background())
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

	settings, err := m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldNatsRequestTimeoutSeconds).Only(context.Background())
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

	settings, err := m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldCountry).Only(context.Background())
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

	settings, err := m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldUserCertYearsValid).Only(context.Background())
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

	settings, err := m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldRefreshTimeInMinutes).Only(context.Background())
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

	settings, err := m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldSessionLifetimeInMinutes).Where(settings.Not(settings.HasTenant())).Only(context.Background())
	if err != nil {
		return 0, err
	}

	return settings.SessionLifetimeInMinutes, nil
}

func (m *Model) UpdateSessionLifetime(settingsId, sessionLifetime int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetSessionLifetimeInMinutes(sessionLifetime).Exec(context.Background())
}

func (m *Model) GetDefaultAgentFrequency(tenantID string) (int, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldAgentReportFrequenceInMinutes).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return 0, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldAgentReportFrequenceInMinutes).Only(context.Background())
	}

	if err != nil {
		return 0, err
	}

	return s.AgentReportFrequenceInMinutes, nil
}

func (m *Model) UpdateAgentFrequency(settingsId, frequency int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetAgentReportFrequenceInMinutes(frequency).Exec(context.Background())
}

func (m *Model) GetDefaultUpdateChannel() (string, error) {
	var err error

	settings, err := m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldUpdateChannel).Only(context.Background())
	if err != nil {
		return "", err
	}

	return settings.UpdateChannel, nil
}

func (m *Model) UpdateRequestVNCPIN(settingsId int, requestPIN bool) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetRequestVncPin(requestPIN).Exec(context.Background())
}

func (m *Model) GetDefaultRequestVNCPIN(tenantID string) (bool, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldRequestVncPin).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return false, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldRequestVncPin).Only(context.Background())
	}

	if err != nil {
		return false, err
	}

	return s.RequestVncPin, nil
}

func (m *Model) UpdateOpenUEMChannel(settingsId int, updateChannel string) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetUpdateChannel(updateChannel).Exec(context.Background())
}

func (m *Model) GetDefaultWingetFrequency(tenantID string) (int, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldProfilesApplicationFrequenceInMinutes).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return 0, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldProfilesApplicationFrequenceInMinutes).Only(context.Background())
	}

	if err != nil {
		return 0, err
	}

	return s.ProfilesApplicationFrequenceInMinutes, nil
}

func (m *Model) UpdateWingetFrequency(settingsId, frequency int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetProfilesApplicationFrequenceInMinutes(frequency).Exec(context.Background())
}

func (m *Model) GetDefaultSFTPDisabled(tenantID string) (bool, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldDisableSftp).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return false, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldDisableSftp).Only(context.Background())
	}

	if err != nil {
		return false, err
	}

	return s.DisableSftp, nil
}

func (m *Model) UpdateSFTPDisabled(settingsId int, disableSFTP bool) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetDisableSftp(disableSFTP).Exec(context.Background())
}

func (m *Model) GetDefaultRemoteAssistanceDisabled(tenantID string) (bool, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldDisableRemoteAssistance).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return false, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldDisableRemoteAssistance).Only(context.Background())
	}

	if err != nil {
		return false, err
	}
	return s.DisableRemoteAssistance, nil
}

func (m *Model) UpdateRemoteAssistanceDisabled(settingsId int, disableRemoteAssistance bool) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetDisableRemoteAssistance(disableRemoteAssistance).Exec(context.Background())
}

func (m *Model) GetDefaultDetectRemoteAgents(tenantID string) (bool, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldDetectRemoteAgents).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return false, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldDetectRemoteAgents).Only(context.Background())
	}

	if err != nil {
		return false, err
	}
	return s.DetectRemoteAgents, nil
}

func (m *Model) UpdateDetectRemoteAgents(settingsId int, detectRemoteAgents bool) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetDetectRemoteAgents(detectRemoteAgents).Exec(context.Background())
}

func (m *Model) GetDefaultAutoAdmitAgents(tenantID string) (bool, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldAutoAdmitAgents).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return false, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldAutoAdmitAgents).Only(context.Background())
	}

	if err != nil {
		return false, err
	}

	return s.AutoAdmitAgents, nil
}

func (m *Model) UpdateAutoAdmitAgents(settingsId int, autoAdmitAgents bool) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetAutoAdmitAgents(autoAdmitAgents).Exec(context.Background())
}

func (m *Model) GetGeneralSettings(tenantID string) (*openuem_ent.Settings, error) {
	var s *openuem_ent.Settings
	var query *openuem_ent.SettingsQuery

	if tenantID == "-1" {
		query = m.Client.Settings.Query().WithTag().Select(
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
			settings.FieldProfilesApplicationFrequenceInMinutes,
			settings.FieldUseWinget,
			settings.FieldUseFlatpak,
			settings.FieldDisableSftp,
			settings.FieldDisableRemoteAssistance,
			settings.FieldDetectRemoteAgents,
			settings.FieldAutoAdmitAgents,
			settings.TagColumn,
		).Where(settings.Not(settings.HasTenantWith()))
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return nil, err
		}

		query = m.Client.Settings.Query().WithTag().Select(
			settings.FieldID,
			settings.FieldAgentReportFrequenceInMinutes,
			settings.FieldRequestVncPin,
			settings.FieldProfilesApplicationFrequenceInMinutes,
			settings.FieldUseWinget,
			settings.FieldUseFlatpak,
			settings.FieldDisableSftp,
			settings.FieldDisableRemoteAssistance,
			settings.FieldDetectRemoteAgents,
			settings.FieldAutoAdmitAgents,
			settings.TagColumn,
		).Where(settings.HasTenantWith(tenant.ID(id)))
	}

	s, err := query.Only(context.Background())
	if err != nil {
		if !openuem_ent.IsNotFound(err) {
			return nil, err
		} else {
			if tenantID == "-1" {
				if err := m.Client.Settings.Create().Exec(context.Background()); err != nil {
					return nil, err
				}
				return query.Only(context.Background())
			} else {
				id, err := strconv.Atoi(tenantID)
				if err != nil {
					return nil, err
				}

				if err := m.CloneGlobalSettings(id); err != nil {
					return nil, err
				}
				return query.Only(context.Background())
			}
		}
	}

	return s, nil
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

func (m *Model) AddAdmittedTag(settingsId int, tag int) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetTagID(tag).Exec(context.Background())
}

func (m *Model) RemoveAdmittedTag(settingsId int) error {
	return m.Client.Settings.UpdateOneID(settingsId).ClearTag().Exec(context.Background())
}

func (m *Model) UpdateUseWinget(settingsId int, useWinGet bool) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetUseWinget(useWinGet).Exec(context.Background())
}

func (m *Model) GetDefaultUseWinget(tenantID string) (bool, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldUseWinget).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return false, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldUseWinget).Only(context.Background())
	}

	if err != nil {
		return false, err
	}

	return s.UseWinget, nil
}

func (m *Model) UpdateUseFlatpak(settingsId int, useFlatpak bool) error {
	return m.Client.Settings.UpdateOneID(settingsId).SetUseFlatpak(useFlatpak).Exec(context.Background())
}

func (m *Model) GetDefaultUseFlatpak(tenantID string) (bool, error) {
	var err error
	var s *openuem_ent.Settings

	if tenantID == "-1" {
		s, err = m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Select(settings.FieldUseFlatpak).Only(context.Background())
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return false, err
		}

		s, err = m.Client.Settings.Query().Where(settings.HasTenantWith(tenant.ID(id))).Select(settings.FieldUseFlatpak).Only(context.Background())
	}

	if err != nil {
		return false, err
	}

	return s.UseFlatpak, nil
}

func (m *Model) CloneGlobalSettings(tenantID int) error {
	s, err := m.Client.Settings.Query().WithTag().Where(settings.Not(settings.HasTenant())).Only(context.Background())
	if err != nil {
		return err
	}

	query := m.Client.Settings.Create().
		SetAgentReportFrequenceInMinutes(s.AgentReportFrequenceInMinutes).
		SetAutoAdmitAgents(s.AutoAdmitAgents).
		SetCountry(s.Country).
		SetDetectRemoteAgents(s.DetectRemoteAgents).
		SetDisableRemoteAssistance(s.DisableRemoteAssistance).
		SetDisableSftp(s.DisableSftp).
		SetMaxUploadSize(s.MaxUploadSize).
		SetNatsRequestTimeoutSeconds(s.NatsRequestTimeoutSeconds).
		SetMessageFrom(s.MessageFrom).
		SetProfilesApplicationFrequenceInMinutes(s.ProfilesApplicationFrequenceInMinutes).
		SetRefreshTimeInMinutes(s.RefreshTimeInMinutes).
		SetRequestVncPin(s.RequestVncPin).
		SetSMTPAuth(s.SMTPAuth).
		SetSMTPPassword(s.SMTPPassword).
		SetSMTPPort(s.SMTPPort).
		SetSMTPServer(s.SMTPServer).
		SetSMTPStarttls(s.SMTPStarttls).
		SetSMTPTLS(s.SMTPTLS).
		SetSMTPUser(s.SMTPUser).
		SetSessionLifetimeInMinutes(s.SessionLifetimeInMinutes).
		SetUpdateChannel(s.UpdateChannel).
		SetUseFlatpak(s.UseFlatpak).
		SetUseWinget(s.UseWinget).
		SetUserCertYearsValid(s.UserCertYearsValid).
		SetTenantID(tenantID)

	if s.Edges.Tag != nil {
		query = query.SetTagID(s.Edges.Tag.ID)
	}

	return query.Exec(context.Background())
}

func (m *Model) ApplyGlobalSettings(tenantID int) error {
	s, err := m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).Only(context.Background())
	if err != nil {
		return err
	}

	query := m.Client.Settings.Update().Where(settings.HasTenantWith(tenant.ID(tenantID))).
		SetAgentReportFrequenceInMinutes(s.AgentReportFrequenceInMinutes).
		SetAutoAdmitAgents(s.AutoAdmitAgents).
		SetCountry(s.Country).
		SetDetectRemoteAgents(s.DetectRemoteAgents).
		SetDisableRemoteAssistance(s.DisableRemoteAssistance).
		SetDisableSftp(s.DisableSftp).
		SetMaxUploadSize(s.MaxUploadSize).
		SetNatsRequestTimeoutSeconds(s.NatsRequestTimeoutSeconds).
		SetMessageFrom(s.MessageFrom).
		SetProfilesApplicationFrequenceInMinutes(s.ProfilesApplicationFrequenceInMinutes).
		SetRefreshTimeInMinutes(s.RefreshTimeInMinutes).
		SetRequestVncPin(s.RequestVncPin).
		SetSMTPAuth(s.SMTPAuth).
		SetSMTPPassword(s.SMTPPassword).
		SetSMTPPort(s.SMTPPort).
		SetSMTPServer(s.SMTPServer).
		SetSMTPStarttls(s.SMTPStarttls).
		SetSMTPTLS(s.SMTPTLS).
		SetSMTPUser(s.SMTPUser).
		SetSessionLifetimeInMinutes(s.SessionLifetimeInMinutes).
		SetUpdateChannel(s.UpdateChannel).
		SetUseFlatpak(s.UseFlatpak).
		SetUseWinget(s.UseWinget).
		SetUserCertYearsValid(s.UserCertYearsValid).
		SetTenantID(tenantID)

	query = query.ClearTag()

	if s.Edges.Tag != nil {
		query = query.SetTagID(s.Edges.Tag.ID)
	}

	return query.Exec(context.Background())
}
