package models

import (
	"context"
	"strconv"

	openuem_ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/settings"
	"github.com/open-uem/ent/tenant"
)

type SMTPSettings struct {
	ID       int
	Server   string
	Port     int
	User     string
	Password string
	Auth     string
	MailFrom string
}

func (m *Model) GetSMTPSettings(tenantID string) (*openuem_ent.Settings, error) {
	var err error
	var s *openuem_ent.Settings

	query := m.Client.Settings.Query().Select(
		settings.FieldSMTPServer,
		settings.FieldSMTPPort,
		settings.FieldSMTPUser,
		settings.FieldSMTPPassword,
		settings.FieldSMTPAuth,
		settings.FieldSMTPTLS,
		settings.FieldSMTPStarttls,
		settings.FieldMessageFrom)

	if tenantID == "-1" {
		s, err = query.Where(settings.Not(settings.HasTenant())).Only(context.Background())
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
	} else {
		id, err := strconv.Atoi(tenantID)
		if err != nil {
			return nil, err
		}

		s, err = query.Where(settings.HasTenantWith(tenant.ID(id))).Only(context.Background())
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
	}

	return s, nil
}

func (m *Model) UpdateSMTPSettings(settings *SMTPSettings) error {
	mainQuery := m.Client.Settings.UpdateOneID(settings.ID).SetSMTPServer(settings.Server).SetSMTPPort(settings.Port).SetSMTPUser(settings.User).SetSMTPPassword(settings.Password).SetMessageFrom(settings.MailFrom)
	return mainQuery.Exec(context.Background())
}

func (m *Model) IsSMTPConfigured() bool {
	s, err := m.Client.Settings.Query().Where(settings.Not(settings.HasTenant())).First(context.Background())
	if err != nil {
		return false
	}

	return s.SMTPServer != "" && s.SMTPPort != 0
}

func (m *Model) GetSMTPPasswords() ([]*openuem_ent.Settings, error) {
	return m.Client.Settings.Query().Select(settings.FieldID, settings.FieldSMTPPassword).All(context.Background())
}

func (m *Model) UpdateSMTPPassword(settingID int, password string) error {
	return m.Client.Settings.UpdateOneID(settingID).SetSMTPPassword(password).Exec(context.Background())
}
