package models

import (
	"context"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/rustdesk"
	"github.com/open-uem/ent/tenant"
)

func (m *Model) GetRustDeskSettings(tenantID int) ([]*ent.RustDesk, error) {
	return m.Client.RustDesk.Query().Where(rustdesk.HasTenantWith(tenant.ID(tenantID))).All(context.Background())
}

func (m *Model) SaveRustDeskSettings(tenantID int, rendezvousServer, relayServer, key, apiServer, whitelist string, useDirectIPAccess, usePermanentPassword bool) error {

	rd, err := m.Client.RustDesk.Query().First(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return m.Client.RustDesk.Create().
				SetCustomRendezvousServer(rendezvousServer).
				SetRelayServer(relayServer).
				SetKey(key).
				SetAPIServer(apiServer).
				SetWhitelist(whitelist).
				SetUsePermanentPassword(usePermanentPassword).
				SetDirectIPAccess(useDirectIPAccess).
				SetTenantID(tenantID).
				Exec(context.Background())
		}
		return err
	}

	return m.Client.RustDesk.UpdateOneID(rd.ID).
		SetCustomRendezvousServer(rendezvousServer).
		SetRelayServer(relayServer).
		SetKey(key).
		SetAPIServer(apiServer).
		SetWhitelist(whitelist).
		SetUsePermanentPassword(usePermanentPassword).
		SetDirectIPAccess(useDirectIPAccess).
		SetTenantID(tenantID).
		Exec(context.Background())
}
