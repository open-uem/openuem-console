package models

import (
	"context"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/rustdesk"
	"github.com/open-uem/ent/tenant"
)

func (m *Model) GetRustDeskSettings(tenantID int) ([]*ent.Rustdesk, error) {
	settings, err := m.GetTenantRustDeskSettings(tenantID)
	if err != nil {
		return nil, err
	}
	if len(settings) == 0 {
		settings, err = m.GetGlobalRustDeskSettings()
		if err != nil {
			return nil, err
		}
		if len(settings) == 0 {
			return []*ent.Rustdesk{}, nil
		}
		return settings, nil
	} else {
		return settings, nil
	}
}

func (m *Model) GetTenantRustDeskSettings(tenantID int) ([]*ent.Rustdesk, error) {
	return m.Client.Rustdesk.Query().Where(rustdesk.HasTenantWith(tenant.ID(tenantID))).All(context.Background())
}

func (m *Model) GetGlobalRustDeskSettings() ([]*ent.Rustdesk, error) {
	return m.Client.Rustdesk.Query().Where(rustdesk.Not(rustdesk.HasTenant())).All(context.Background())
}

func (m *Model) SaveRustDeskSettings(tenantID int, rendezvousServer, relayServer, key, apiServer, whitelist string, useDirectIPAccess, usePermanentPassword bool) error {
	var rd *ent.Rustdesk
	var err error

	if tenantID != -1 {
		rd, err = m.Client.Rustdesk.Query().Where(rustdesk.HasTenantWith(tenant.ID(tenantID))).First(context.Background())
	} else {
		rd, err = m.Client.Rustdesk.Query().Where(rustdesk.Not(rustdesk.HasTenant())).First(context.Background())
	}

	if err != nil {
		if ent.IsNotFound(err) {
			query := m.Client.Rustdesk.Create().
				SetCustomRendezvousServer(rendezvousServer).
				SetRelayServer(relayServer).
				SetKey(key).
				SetAPIServer(apiServer).
				SetWhitelist(whitelist).
				SetUsePermanentPassword(usePermanentPassword).
				SetDirectIPAccess(useDirectIPAccess)

			if tenantID != -1 {
				query.AddTenantIDs(tenantID)
			}

			return query.Exec(context.Background())
		}
		return err
	}

	return m.Client.Rustdesk.UpdateOneID(rd.ID).
		SetCustomRendezvousServer(rendezvousServer).
		SetRelayServer(relayServer).
		SetKey(key).
		SetAPIServer(apiServer).
		SetWhitelist(whitelist).
		SetUsePermanentPassword(usePermanentPassword).
		SetDirectIPAccess(useDirectIPAccess).
		Exec(context.Background())
}

func (m *Model) HasRustDeskSettings(tenantID int) bool {
	s, err := m.GetRustDeskSettings(tenantID)
	if err == nil && len(s) != 0 {
		return true
	}

	return false
}
