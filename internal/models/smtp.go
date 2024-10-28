package models

import (
	"context"

	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/settings"
)

func (m *Model) GetSMTPSettings() (*openuem_ent.Settings, error) {
	return m.Client.Settings.Query().Select(
		settings.FieldSMTPServer,
		settings.FieldSMTPPort,
		settings.FieldSMTPUser,
		settings.FieldSMTPPassword,
		settings.FieldSMTPAuth,
		settings.FieldSMTPTLS,
		settings.FieldSMTPStarttls,
		settings.FieldMessageFrom,
	).Only(context.Background())
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
