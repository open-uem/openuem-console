package models

import (
	"context"
	"errors"

	openuem_ent "github.com/open-uem/ent"
	"github.com/sethvargo/go-password/password"
)

func (m *Model) GetAuthenticationSettings() (*openuem_ent.Authentication, error) {

	settings, err := m.Client.Authentication.Query().Only(context.Background())
	if err != nil {
		if !openuem_ent.IsNotFound(err) {
			return nil, err
		}

		return m.Client.Authentication.Create().Save(context.Background())
	}

	return settings, nil
}

func (m *Model) SaveAuthenticationSettings(useCertificates bool, allowRegister bool, useOIDC bool, provider string,
	server string, clientID string, roleAdmin string, roleOperator string, roleUser string, autoCreate bool, autoApprove bool, usePasswd bool) error {

	s, err := m.Client.Authentication.Query().Only(context.Background())
	if err != nil {
		return err
	}

	update := m.Client.Authentication.UpdateOneID(s.ID).
		SetUsePasswd(usePasswd).
		SetUseCertificates(useCertificates).
		SetAllowRegister(allowRegister).
		SetUseOIDC(useOIDC).
		SetOIDCProvider(provider).
		SetOIDCIssuerURL(server).
		SetOIDCClientID(clientID).
		SetOIDCRoleAdmin(roleAdmin).
		SetOIDCRoleOperator(roleOperator).
		SetOIDCRoleUser(roleUser).
		SetOIDCAutoCreateAccount(autoCreate).
		SetOIDCAutoApprove(autoApprove)

	// Create encryption key for OIDC cookie
	if useOIDC {
		if s.OIDCCookieEncriptionKey == "" {
			key, err := password.Generate(32, 10, 0, false, true)
			if err != nil {
				return errors.New("could not generate the cookie encryption key")
			}
			update.SetOIDCCookieEncriptionKey(key)
		}
	} else {
		update.SetOIDCCookieEncriptionKey("")
	}

	return update.Exec(context.Background())
}

func (m *Model) ReEnableCertificatesAuth() error {

	s, err := m.Client.Authentication.Query().Only(context.Background())
	if err != nil {
		return err
	}

	return m.Client.Authentication.UpdateOneID(s.ID).SetUseCertificates(true).Exec(context.Background())
}

func (m *Model) ReEnablePasswdAuth() error {

	s, err := m.Client.Authentication.Query().Only(context.Background())
	if err != nil {
		return err
	}

	return m.Client.Authentication.UpdateOneID(s.ID).SetUsePasswd(true).Exec(context.Background())
}

func (m *Model) IsPasswdAuthEnabled() bool {
	s, err := m.Client.Authentication.Query().Only(context.Background())
	if err != nil {
		return false
	}

	return s.UsePasswd
}
