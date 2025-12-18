package models

import (
	"context"
	"strings"

	"github.com/open-uem/ent"
	"github.com/open-uem/ent/agent"
	"github.com/open-uem/ent/netbird"
	"github.com/open-uem/ent/netbirdsettings"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/nats"
)

func (m *Model) GetNetbirdSettings(tenantID int) (*ent.NetbirdSettings, error) {
	s, err := m.Client.NetbirdSettings.Query().Where(netbirdsettings.HasTenantWith(tenant.ID(tenantID))).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return m.Client.NetbirdSettings.Create().AddTenantIDs(tenantID).Save(context.Background())
		}

		return nil, err
	}

	return s, nil
}

func (m *Model) SaveNetbirdSettings(tenantID int, managementURL string, accessToken string) error {
	var err error

	nb, err := m.Client.NetbirdSettings.Query().Where(netbirdsettings.HasTenantWith(tenant.ID(tenantID))).First(context.Background())

	if err != nil {
		if ent.IsNotFound(err) {
			return m.Client.NetbirdSettings.Create().
				SetManagementURL(managementURL).
				SetAccessToken(accessToken).
				AddTenantIDs(tenantID).Exec(context.Background())
		}
		return err
	}

	return m.Client.NetbirdSettings.UpdateOneID(nb.ID).
		SetManagementURL(managementURL).
		SetAccessToken(accessToken).
		Exec(context.Background())
}

func (m *Model) SaveNetbirdInfo(agentID string, data nats.Netbird) error {
	return m.Client.Netbird.
		Create().
		SetVersion(data.Version).
		SetInstalled(data.Installed).
		SetIP(data.IP).
		SetSSHEnabled(data.SSHEnabled).
		SetProfile(data.Profile).
		SetManagementConnected(data.ManagementConnected).
		SetManagementURL(data.ManagementURL).
		SetSignalConnected(data.SignalConnected).
		SetSignalURL(data.SignalURL).
		SetPeersConnected(data.PeersConnected).
		SetPeersTotal(data.PeersTotal).
		SetServiceStatus(data.ServiceStatus).
		SetProfilesAvailable(strings.Join(data.Profiles, ",")).
		SetDNSServer(strings.Join(data.DNSServers, ",")).
		SetOwnerID(agentID).
		OnConflictColumns(netbird.OwnerColumn).
		UpdateNewValues().
		Exec(context.Background())
}

func (m *Model) SetNetbirdAsUninstalled(agentID string) error {
	return m.Client.Netbird.Update().SetInstalled(false).Where(netbird.HasOwnerWith(agent.ID(agentID))).Exec(context.Background())
}
