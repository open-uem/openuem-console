package models

import (
	"context"

	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/settings"
)

func (m *Model) GetSMTPSettings() (*openuem_ent.Settings, error) {

	query := m.Client.Settings.Query().Select(
		settings.FieldSMTPServer,
		settings.FieldSMTPPort,
		settings.FieldSMTPUser,
		settings.FieldSMTPPassword,
		settings.FieldSMTPAuth,
		settings.FieldSMTPTLS,
		settings.FieldSMTPStarttls,
		settings.FieldMessageFrom)

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

func (m *Model) UpdateSMTPSettings(settings *SMTPSettings) error {
	mainQuery := m.Client.Settings.UpdateOneID(settings.ID).SetSMTPServer(settings.Server).SetSMTPPort(settings.Port).SetSMTPUser(settings.User).SetSMTPPassword(settings.Password).SetMessageFrom(settings.MailFrom)
	return mainQuery.Exec(context.Background())
}

type SMTPSettings struct {
	ID       int
	Server   string
	Port     int
	User     string
	Password string
	Auth     string
	MailFrom string
}
